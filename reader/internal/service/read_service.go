package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"github.com/Azure/azure-sdk-for-go/sdk/messaging/azeventhubs"
	"log"
	azure2 "reader/internal/repository/azure"
	"reader/internal/repository/mysql"
	"reader/internal/service/model"
	"sync"
	"time"
)

type ReadService struct {
	EventHubConsumerProcessor *azure2.EventHubConsumerProcessor
	EventHubProducer          *azure2.EventHubProducer
	MysqlRepository           *mysql.MysqlRepository
}

func (r *ReadService) Run(ctx context.Context) error {
	// 処理順番
	// done 1. event hub preからeventを取得する
	// 2. 1で取得したeventのspot_idでRDBをupsert
	// 3. RDBからis_vespa_updated: falseのレコードを取得
	// 4. 2.3で取得した複数spot_idからspotデータを取得する
	// 5. event hub postにeventを投げる

	// consumerProcessorを起動
	err := r.EventHubConsumerProcessor.Run(ctx)
	if err != nil {
		return err
	}

	var wg sync.WaitGroup
	// NOTE: 同時に起動するpartitionClientの数を3台に制御
	ch := make(chan struct{}, 3)
	for {
		// NOTE: partitionから要求があるたびに、そのpartitionに対するClientが作成される
		processorPartitionClient := r.EventHubConsumerProcessor.DispatchPartitionClients(ctx)
		if processorPartitionClient == nil {
			break
		}
		// NOTE: clientが作成されたらchannelのbufferを１つ満たす
		// channelのbufferが3なので、4つめからはここで処理待ちが発生する
		// eventHub -> vepsa upsertを処理をアプリケーションで制御する場合は
		// このパターンを利用する
		ch <- struct{}{}

		log.Printf("partitionClient(%s) is running", processorPartitionClient.PartitionID())

		// NOTE: clientごとに別のgoroutineでreceive処理を行う
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := r.processEventsForPartition(processorPartitionClient, ctx)
			if err != nil {
				log.Printf("error process eventhub partitionId %s: %s", processorPartitionClient.PartitionID(), err.Error())
			}
			log.Println("子goroutine終了")
			// NOTE: channel内のbufferを１つ解放
			<-ch
		}()
	}
	wg.Wait()

	return nil
}

func (r *ReadService) processEventsForPartition(partitionClient *azeventhubs.ProcessorPartitionClient, ctx context.Context) error {
	// closure
	// 実際に実行されるまでshutdownPartitionResourceの引数は評価されない
	defer func() {
		shutdownPartitionResource(partitionClient, ctx)
	}()

parentLoop:
	for {
		select {
		case <-ctx.Done():
			break parentLoop
		default:
			receiveCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
			// NOTE: 最大100件取得完了する、5秒経過するまで待ち受ける
			// 5秒経過した場合、5秒感で取得できたeventを返す
			events, err := partitionClient.ReceiveEvents(receiveCtx, 100, nil)

			// ↑が終了したらとりあえずcontext.WithTimeoutで利用していたresource等を閉じる
			cancel()

			// NOTE: errorならprocessEventsForPartition関数から抜ける
			if err != nil || errors.Is(err, context.DeadlineExceeded) {
				var eventhubError *azeventhubs.Error
				if errors.As(err, &eventhubError) && eventhubError.Code == azeventhubs.ErrorCodeOwnershipLost {
					log.Printf("decode event error: %s", err.Error())
					return err
				}
			}

			if len(events) == 0 {
				continue
			}

			log.Printf("partitionId: %s receive %d events", partitionClient.PartitionID(), len(events))

			// NOTE: errorならprocessEventsForPartition関数から抜ける
			event, err := decodeEvent(events)
			if err != nil {
				log.Printf("decode event error: %s", err.Error())
				return err
			}

			// TODO: 後続処理実装

			r.updateRdb(event)
			r.getSpotIdFromRdb()
			r.getSpotDataFromFirestore()
			r.sendEventTo()

			if err == nil {
				// NOTE: 正常に処理された場合はcheckpointを更新する
				if err := partitionClient.UpdateCheckpoint(ctx, events[len(events)-1], nil); err != nil {
					break parentLoop
				}
			}
		}

	}
	return nil
}

func decodeEvent(events []*azeventhubs.ReceivedEventData) ([]model.PreEventData, error) {
	preEventDataArray := make([]model.PreEventData, len(events))
	for i := range events {
		var buf bytes.Buffer
		buf.Write(events[i].Body)
		var preEventData model.PreEventData
		err := json.NewDecoder(&buf).Decode(&preEventData)
		if err != nil {
			return nil, err
		}
		preEventDataArray[i] = preEventData
	}

	return preEventDataArray, nil
}

func (r *ReadService) updateRdb(events []model.PreEventData) {
}

func (r *ReadService) getSpotIdFromRdb() {

}

func (r *ReadService) getSpotDataFromFirestore() {

}

func (r *ReadService) sendEventTo() {

}

func shutdownPartitionResource(partitionClient *azeventhubs.ProcessorPartitionClient, ctx context.Context) {
	if partitionClient != nil {
		log.Printf("shutdown partitionClient(%s)", partitionClient.PartitionID())
		partitionClient.Close(ctx)
	}

}
