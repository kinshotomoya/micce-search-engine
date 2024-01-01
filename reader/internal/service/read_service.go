package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Azure/azure-sdk-for-go/sdk/messaging/azeventhubs"
	"reader/internal"
	model3 "reader/internal/domain/model"
	azure2 "reader/internal/repository/azure"
	"reader/internal/repository/firestore"
	"reader/internal/repository/mysql"
	model2 "reader/internal/repository/mysql/model"
	"reader/internal/service/model"
	"sync"
	"time"
)

type ReadService struct {
	EventHubConsumerProcessor *azure2.EventHubConsumerProcessor
	EventHubProducer          *azure2.EventHubProducer
	MysqlRepository           *mysql.MysqlRepository
	CustomTime                model3.CustomTimeInterface
	FireStoreClient           *firestore.FireStoreClient
}

func (r *ReadService) Run(signalCtx context.Context) error {
	// 処理順番
	// done 1. event hub preからeventを取得する
	// 2. 1で取得したeventのspot_idでRDBをupsert
	// 3. RDBからis_vespa_updated: falseのレコードを取得
	// 4. 2.3で取得した複数spot_idからspotデータを取得する
	// 5. event hub postにeventを投げる

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

		// NOTE: clientごとに別のgoroutineでreceive処理を行う
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
	event, err := decodeEvent(events)
	if err != nil {
		internal.Logger.Error(fmt.Sprintf("decode event error: %s", err.Error()))
		return err
	}

	updateUpdateProcessCtx, updateUpdateProcessCtxCancel := context.WithTimeout(mainCtx, 3*time.Second)
	defer updateUpdateProcessCtxCancel()
	spotIdsToUpdate, err := r.updateUpdateProcess(updateUpdateProcessCtx, event)
	if err != nil {
		spotIds := make([]string, len(event))
		for i := range event {
			spotIds[i] = event[i].SpotId
		}
		internal.Logger.Error(fmt.Sprintf("fatal update %s: %s", spotIds, err.Error()))
		return err
	}

	getSpotDataFromFirestoreCtx, getSpotDataFromFirestoreCtxCancel := context.WithTimeout(mainCtx, 3*time.Second)
	defer getSpotDataFromFirestoreCtxCancel()
	err = r.getSpotDataFromFirestore(getSpotDataFromFirestoreCtx, spotIdsToUpdate)
	if err != nil {
		internal.Logger.Error(fmt.Sprintf("fatal get spot data from firestore: %s", err.Error()))
		return err
	}

	if err == nil {
		// NOTE: 正常に処理された場合はcheckpointを更新する
		updateCheckpointCtx, updateCheckpointCtxCancel := context.WithTimeout(mainCtx, 3*time.Second)
		defer updateCheckpointCtxCancel()
		if err := partitionClient.UpdateCheckpoint(updateCheckpointCtx, events[len(events)-1], nil); err != nil {
			return err
		}
		internal.Logger.Info(fmt.Sprintf("partitionClient(%s): update checkpoint", partitionClient.PartitionID()))
	}
	return nil
}

func decodeEvent(events []*azeventhubs.ReceivedEventData) ([]model.PreEventData, error) {
	preEventDataArray := make([]model.PreEventData, len(events))
	for i := range events {
		internal.Logger.Info(string(events[i].Body))
		var buf bytes.Buffer
		buf.Write(events[i].Body)
		var preEventData model.PreEventData
		err := json.NewDecoder(&buf).Decode(&preEventData)
		if err != nil {
			internal.Logger.Error(fmt.Sprintf("fatal decode event data: %s", err.Error()))
			return nil, err
		}
		preEventDataArray[i] = preEventData
	}

	return preEventDataArray, nil
}

func (r *ReadService) updateUpdateProcess(ctx context.Context, events []model.PreEventData) ([]string, error) {
	conditions := make([]model2.UpsertCondition, len(events))
	for i := range events {
		conditions[i] = model2.UpsertCondition{
			SpotId:         events[i].SpotId,
			UpdatedAt:      r.CustomTime.DatetimeNow(),
			VespaUpdatedAt: nil,
			IsVespaUpdated: false,
		}
	}

	spotIdsToUpdate, err := r.MysqlRepository.UpsertIsVespaUpdatedAndGetSpotIdsToUpdate(ctx, conditions)
	if err != nil {
		return nil, err
	}

	return spotIdsToUpdate, nil
}

func (r *ReadService) getSpotDataFromFirestore(ctx context.Context, spotIdsToUpdate []string) error {
	documentIter := r.FireStoreClient.GetDocumentsBySpotIds(ctx, spotIdsToUpdate)
	eventDatas := make([]azeventhubs.EventData, 0, len(spotIdsToUpdate))
	for {
		snapShot, err := documentIter.Next()
		if err != nil {
			internal.Logger.Info("もうiteratorにないのでloop抜ける")
			break
		}
		doc := firestore.CreateDocument(snapShot)
		var buf bytes.Buffer
		err = json.NewEncoder(&buf).Encode(doc)
		if err != nil {
			internal.Logger.Error(fmt.Sprintf("fatal encode firestore document: %s", err.Error()))
			return err
		}
		eventData := azeventhubs.EventData{
			Body: buf.Bytes(),
		}
		eventDatas = append(eventDatas, eventData)
	}

	err := r.sendToEventHub(ctx, eventDatas)
	if err != nil {
		internal.Logger.Error(fmt.Sprintf("fatal send data to eventhub: %s", err.Error()))
		return err
	}
	internal.Logger.Info(fmt.Sprintf("send %d count event data to eventhub", len(eventDatas)))
	return nil

}
func (r *ReadService) sendToEventHub(ctx context.Context, eventDatas []azeventhubs.EventData) error {
	timeoutCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	eventDataBatch, err := r.EventHubProducer.CreateEventBatch(timeoutCtx, eventDatas)
	if err != nil {
		internal.Logger.Error(fmt.Sprintf("fatal create event data batch: %s", err.Error()))
		return err
	}
	err = r.EventHubProducer.Send(timeoutCtx, eventDataBatch)
	if err != nil {
		internal.Logger.Error(fmt.Sprintf("fatal send event data batch: %s", err.Error()))
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
