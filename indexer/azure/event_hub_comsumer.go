package azure

import (
	"github.com/Azure/azure-sdk-for-go/sdk/messaging/azeventhubs"
	"github.com/Azure/azure-sdk-for-go/sdk/messaging/azeventhubs/checkpoints"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/container"
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
