package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Azure/azure-sdk-for-go/sdk/messaging/azeventhubs"
	"indexer/internal"
	model3 "indexer/internal/domain/model"
	azure2 "indexer/internal/repository/azure"
	"indexer/internal/repository/mysql"
	model2 "indexer/internal/repository/mysql/model"
	"indexer/internal/repository/vespa"
	"sync"
	"time"
)

type ReadService struct {
	EventHubConsumerProcessor *azure2.EventHubConsumerProcessor
	VespaClient               *vespa.VespaClient
	MysqlRepository           *mysql.MysqlRepository
	CustomTime                model3.CustomTimeInterface
}

func (r *ReadService) Run(signalCtx context.Context) error {
	var wg sync.WaitGroup
	// NOTE: 同時に起動するpartitionClientの数を3台に制御
	ch := make(chan struct{}, 3)
	for {
		// NOTE: partitionから要求があるたびに、そのpartitionに対するClientが作成される
		processorPartitionClient := r.EventHubConsumerProcessor.DispatchPartitionClients(signalCtx)
		if processorPartitionClient == nil {
			break
		}
		ch <- struct{}{}

		internal.Logger.Info(fmt.Sprintf("partitionClient(%s) is running", processorPartitionClient.PartitionID()))

		wg.Add(1)
		go func() {
			defer wg.Done()
			err := r.processEventsForPartition(signalCtx, processorPartitionClient)
			if err != nil {
				internal.Logger.Error(fmt.Sprintf("error process eventhub partitionId %s: %s", processorPartitionClient.PartitionID(), err.Error()))
			}
			internal.Logger.Info(fmt.Sprintf("partitionClient(%s) finished", processorPartitionClient.PartitionID()))
			// NOTE: channel内のbufferを１つ解放
			<-ch
		}()
	}
	wg.Wait()

	return nil
}

func (r *ReadService) processEventsForPartition(signalCtx context.Context, partitionClient *azeventhubs.ProcessorPartitionClient) error {

	// closure
	// 実際に実行されるまでshutdownPartitionResourceの引数は評価されない
	defer func() {
		shutdownPartitionResource(partitionClient, signalCtx)
	}()

parentLoop:
	for {
		select {
		case <-signalCtx.Done():
			break parentLoop
		default:
			err := r.runLoop(partitionClient)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (r *ReadService) runLoop(partitionClient *azeventhubs.ProcessorPartitionClient) error {
	mainCtx, mainCancel := context.WithTimeout(context.Background(), 12*time.Second)
	defer mainCancel()
	// NOTE: 最大100件取得完了する、3秒経過するまで待ち受ける
	// 5秒経過した場合、3秒間で取得できたeventを返す
	receiveCtx, receiveCtxCancel := context.WithTimeout(mainCtx, 3*time.Second)
	defer receiveCtxCancel()
	events, err := partitionClient.ReceiveEvents(receiveCtx, 100, nil)

	// NOTE: errorならprocessEventsForPartition関数から抜ける
	if err != nil || errors.Is(err, context.DeadlineExceeded) {
		var eventhubError *azeventhubs.Error
		if errors.As(err, &eventhubError) && eventhubError.Code == azeventhubs.ErrorCodeOwnershipLost {
			internal.Logger.Error(fmt.Sprintf("event timeout error: %s", err.Error()))
			return err
		}
	}

	if len(events) == 0 {
		internal.Logger.Info(fmt.Sprintf("partitionClient(%s) receive event count is 0", partitionClient.PartitionID()))
		return nil
	}

	internal.Logger.Info(fmt.Sprintf("partitionClient(%s) receive %d events", partitionClient.PartitionID(), len(events)))

	// NOTE: errorならprocessEventsForPartition関数から抜ける
	documents, err := decodeEvent(events)
	if err != nil {
		internal.Logger.Error(fmt.Sprintf("decode event error: %s", err.Error()))
		return err
	}

	err = r.upsertVespa(documents)
	if err != nil {
		internal.Logger.Error(fmt.Sprintf("fatal upsert vespa: %s", err.Error()))
		return err
	}

	updateEventProcessCtx, updateEventProcessCancel := context.WithTimeout(mainCtx, 3*time.Second)
	defer updateEventProcessCancel()
	err = r.updateEventProcess(updateEventProcessCtx, documents)

	if err != nil {
		internal.Logger.Error(fmt.Sprintf("fatal upsert update_process table: %s", err.Error()))
		return err
	}
	internal.Logger.Info("updated update_process table")

	if err == nil {
		// NOTE: 正常に処理された場合はcheckpointを更新する
		updateCheckpointCtx, updateCheckpointCtxCancel := context.WithTimeout(mainCtx, 3*time.Second)
		defer updateCheckpointCtxCancel()
		if err = partitionClient.UpdateCheckpoint(updateCheckpointCtx, events[len(events)-1], nil); err != nil {
			return err
		}
		internal.Logger.Info(fmt.Sprintf("partitionClient(%s): update checkpoint", partitionClient.PartitionID()))
	}
	return nil
}

func decodeEvent(events []*azeventhubs.ReceivedEventData) ([]vespa.Document, error) {
	preEventDataArray := make([]vespa.Document, len(events))
	for i := range events {
		internal.Logger.Info(string(events[i].Body))
		var buf bytes.Buffer
		buf.Write(events[i].Body)
		var document vespa.Document
		err := json.NewDecoder(&buf).Decode(&document)
		if err != nil {
			internal.Logger.Error(fmt.Sprintf("fatal decode event data: %s", err.Error()))
			return nil, err
		}
		preEventDataArray[i] = document
	}

	return preEventDataArray, nil
}

func (r *ReadService) upsertVespa(documents []vespa.Document) error {
	var errs error
	for i := range documents {
		err := r.VespaClient.Upsert(documents[i])
		if err != nil {
			errs = errors.Join(errs, err)
		}
	}
	if errs != nil {
		return errs
	}
	return nil
}

func (r *ReadService) updateEventProcess(ctx context.Context, documents []vespa.Document) error {
	conditions := make([]model2.UpsertCondition, len(documents))
	for i := range documents {
		conditions[i] = model2.UpsertCondition{
			SpotId:         documents[i].Id,
			UpdatedAt:      r.CustomTime.DatetimeNow(),
			VespaUpdatedAt: r.CustomTime.DatetimeNow(),
			IsVespaUpdated: true,
		}
	}

	err := r.MysqlRepository.UpsertIsVespaUpdated(ctx, conditions)
	if err != nil {
		return err
	}

	return nil
}

func shutdownPartitionResource(partitionClient *azeventhubs.ProcessorPartitionClient, ctx context.Context) {
	if partitionClient != nil {
		internal.Logger.Info(fmt.Sprintf("shutdown partitionClient(%s)", partitionClient.PartitionID()))
		partitionClient.Close(ctx)
	}

}
