package azure

import (
	"context"
	"github.com/Azure/azure-sdk-for-go/sdk/messaging/azeventhubs"
	"github.com/Azure/azure-sdk-for-go/sdk/messaging/azeventhubs/checkpoints"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/container"
	"time"
)

const eventHubName = "micce-search-engine"
const storageContainerName = "event-hub-checkpoint"

type EventHubConsumerProcessor struct {
	Processor *azeventhubs.Processor
}

func NewPreEventHubConsumerClient(azureEventHubConnectionString string, azureStorageAccountConnectionString string) (*EventHubConsumerProcessor, error) {
	azBlobClient, err := container.NewClientFromConnectionString(azureStorageAccountConnectionString, storageContainerName, nil)
	if err != nil {
		return nil, err
	}

	checkPointStore, err := checkpoints.NewBlobStore(azBlobClient, nil)
	if err != nil {
		return nil, err
	}

	consumerClient, err := azeventhubs.NewConsumerClientFromConnectionString(azureEventHubConnectionString, eventHubName, azeventhubs.DefaultConsumerGroup, nil)

	if err != nil {
		return nil, err
	}

	option := &azeventhubs.ProcessorOptions{
		// NOTE: 15秒ごとにprocessorがpartitionに問い合わせる
		UpdateInterval: 15 * time.Second,
	}
	processor, err := azeventhubs.NewProcessor(consumerClient, checkPointStore, option)
	if err != nil {
		return nil, err
	}

	return &EventHubConsumerProcessor{
		Processor: processor,
	}, nil

}

func (e *EventHubConsumerProcessor) Run(ctx context.Context) error {
	err := e.Processor.Run(ctx)
	if err != nil {
		return err
	}
	return nil
}

func (e *EventHubConsumerProcessor) DispatchPartitionClients(ctx context.Context) *azeventhubs.ProcessorPartitionClient {
	processorPartitionClient := e.Processor.NextPartitionClient(ctx)
	return processorPartitionClient
}
