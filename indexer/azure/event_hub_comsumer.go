package azure

import (
	"context"
	"errors"
	"fmt"
	"github.com/Azure/azure-sdk-for-go/sdk/messaging/azeventhubs"
	"github.com/Azure/azure-sdk-for-go/sdk/messaging/azeventhubs/checkpoints"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/container"
	"log"
	"time"
)

const eventHubNameSpace = "micce-search-engine"
const eventHubName = "micce-search-engine"
const partitionId = "0"
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
	// TODO: 要調整
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

func DispatchPartitionClients(processor *azeventhubs.Processor, ctx context.Context, ch chan<- string) {
	for {
		// NOTE: partitionから要求があるたびに、そのpartitionに対するClientが作成される
		processorPartitionClient := processor.NextPartitionClient(ctx)
		if processorPartitionClient == nil {
			break
		}

		log.Printf("partitionClient(%s) is running", processorPartitionClient.PartitionID())

		// NOTE: clientごとに別のgoroutineでreceive処理を行う
		go func() {
			err := processEventsForPartition(processorPartitionClient, ctx, ch)
			if err != nil {
				log.Printf("error process eventhub partitionId %s: %s", processorPartitionClient.PartitionID(), err.Error())
				// この時にも終了シグナルをちゃんと受け取れるように
				select {
				case <-ctx.Done():
					shutdownPartitionResource(processorPartitionClient, ctx)
					ch <- fmt.Sprintf("running partitionClient goroutine finished between %s", err.Error())
					break
				}
				return
			}
		}()
	}
}

func processEventsForPartition(partitionClient *azeventhubs.ProcessorPartitionClient, ctx context.Context, ch chan<- string) error {

	// closure
	// 実際に実行されるまでshutdownPartitionResourceの引数は評価されない
	defer func() {
		shutdownPartitionResource(partitionClient, ctx)
	}()

parentLoop:
	for {
		select {
		case <-ctx.Done():
			shutdownPartitionResource(partitionClient, ctx)
			ch <- "running partitionClient goroutine finished"
			break parentLoop
		default:
			receiveCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
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
				// TODO: 実際に受け取ったメッセージをvespaに投げる
				log.Printf("%v", event.Body)
			}

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
