package service

import (
	"context"
	"errors"
	"github.com/Azure/azure-sdk-for-go/sdk/messaging/azeventhubs"
	"log"
	"reader/internal/azure"
	"sync"
	"time"
)

type ReadService struct {
	EventHubConsumerProcessor *azure.EventHubConsumerProcessor
	EventHubProducer          *azure.EventHubProducer
}

func (r ReadService) Run(ctx context.Context) {
	// 処理順番
	// 1. event hub preからeventを取得する
	// 2. 1で取得したeventのspot_idでRDBをupsert
	// 3. RDBからis_vespa_updated: falseのレコードを取得
	// 4. 2.3で取得した複数spot_idからspotデータを取得する
	// 5. event hub postにeventを投げる

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
			err := processEventsForPartition(processorPartitionClient, ctx)
			if err != nil {
				log.Printf("error process eventhub partitionId %s: %s", processorPartitionClient.PartitionID(), err.Error())
			}
			log.Println("子goroutine終了")
			// NOTE: channel内のbufferを１つ解放
			<-ch
		}()
	}
	wg.Wait()

}

func processEventsForPartition(partitionClient *azeventhubs.ProcessorPartitionClient, ctx context.Context) error {
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

			// NOTE: timeout等のエラーの場合はエラーを返す
			if err != nil && errors.Is(err, context.DeadlineExceeded) {
				var eventhubError *azeventhubs.Error

				if errors.As(err, &eventhubError) && eventhubError.Code == azeventhubs.ErrorCodeOwnershipLost {
					return nil
				}
				return err
			}

			if len(events) == 0 {
				continue
			}

			log.Printf("partitionId: %s receive %d events", partitionClient.PartitionID(), len(events))

			// TODO: 後続処理実装
			updateRdb()
			getSpotIdFromRdb()
			getSpotDataFromFirestore()
			sendEventTo()

			if err == nil {
				// NOTE: checkpointを更新する
				if err := partitionClient.UpdateCheckpoint(ctx, events[len(events)-1], nil); err != nil {
					return err
				}
			}
		}

	}
	return nil
}

func updateRdb() {

}

func getSpotIdFromRdb() {

}

func getSpotDataFromFirestore() {

}

func sendEventTo() {

}

func shutdownPartitionResource(partitionClient *azeventhubs.ProcessorPartitionClient, ctx context.Context) {
	if partitionClient != nil {
		log.Printf("shutdown partitionClient(%s)", partitionClient.PartitionID())
		partitionClient.Close(ctx)
	}

}
