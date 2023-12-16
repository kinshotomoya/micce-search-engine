package azure

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"github.com/Azure/azure-sdk-for-go/sdk/messaging/azeventhubs"
	"github.com/Azure/azure-sdk-for-go/sdk/messaging/azeventhubs/checkpoints"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/container"
	vespa2 "indexer/internal/vespa"
	"log"
	"sync"
	"time"
)

const eventHubName = "micce-search-engine"
const storageContainerName = "event-hub-checkpoint"

func NewEventHubConsumerClient(azureEventHubConnectionString string, azureStorageAccountConnectionString string) (*azeventhubs.ConsumerClient, azeventhubs.CheckpointStore, error) {
	azBlobClient, err := container.NewClientFromConnectionString(azureStorageAccountConnectionString, storageContainerName, nil)
	if err != nil {
		return nil, nil, err
	}

	checkPointStore, err := checkpoints.NewBlobStore(azBlobClient, nil)

	if err != nil {
		return nil, nil, err
	}

	consumerClient, err := azeventhubs.NewConsumerClientFromConnectionString(azureEventHubConnectionString, eventHubName, azeventhubs.DefaultConsumerGroup, nil)

	if err != nil {
		return nil, nil, err
	}

	return consumerClient, checkPointStore, nil

}

func NewProcessor(eventHubConsumer *azeventhubs.ConsumerClient, checkpointStore azeventhubs.CheckpointStore) (*azeventhubs.Processor, error) {
	option := &azeventhubs.ProcessorOptions{
		// NOTE: 5秒ごとにprocessorがpartitionに問い合わせる
		UpdateInterval: 5 * time.Second,
	}
	processor, err := azeventhubs.NewProcessor(eventHubConsumer, checkpointStore, option)
	if err != nil {
		return nil, err
	}
	return processor, nil

}

func DispatchPartitionClients(processor *azeventhubs.Processor, ctx context.Context, vespaClient *vespa2.VespaClient) {
	var wg sync.WaitGroup
	// NOTE: 同時に起動するpartitionClientの数を3台に制御
	ch := make(chan struct{}, 3)
	for {
		// NOTE: partitionから要求があるたびに、そのpartitionに対するClientが作成される
		processorPartitionClient := processor.NextPartitionClient(ctx)
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
			err := processEventsForPartition(processorPartitionClient, ctx, vespaClient, &wg)
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

func processEventsForPartition(partitionClient *azeventhubs.ProcessorPartitionClient, ctx context.Context, vespaClient *vespa2.VespaClient, wg *sync.WaitGroup) error {

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
			// NOTE: 100件取得完了する、5秒経過するまで待ち受ける
			// 5秒経過した場合errにErrorが入る
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

			log.Printf("receive %d events", len(events))

			for _, event := range events {
				var buf bytes.Buffer
				buf.Write(event.Body)
				var document vespa2.Document
				err := json.NewDecoder(&buf).Decode(&document)
				if err != nil {
					log.Printf("fatal decode event hub message: %s", err.Error())
					continue
				}
				vespaClient.Upsert(document)

			}
			// TODO: err（eventhubとの接続失敗、vespaへのupsert失敗）の場合には、checkpoint更新しないように

			// NOTE: checkpointを更新する
			if err := partitionClient.UpdateCheckpoint(ctx, events[len(events)-1], nil); err != nil {
				return err
			}
		}

	}
	return nil
}

func shutdownPartitionResource(partitionClient *azeventhubs.ProcessorPartitionClient, ctx context.Context) {
	if partitionClient != nil {
		log.Printf("shutdown partitionClient(%s)", partitionClient.PartitionID())
		partitionClient.Close(ctx)
	}

}
