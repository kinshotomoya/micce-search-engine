package azure

import (
	"context"
	"github.com/Azure/azure-sdk-for-go/sdk/messaging/azeventhubs"
	"reader/internal"
)

// 参考：
// https://learn.microsoft.com/ja-jp/azure/event-hubs/event-hubs-go-get-started-send

// azure console画面からメッセージを確認する方法
// https://learn.microsoft.com/ja-jp/azure/event-hubs/event-hubs-get-connection-string
type EventHubProducer struct {
	producer *azeventhubs.ProducerClient
	option   *azeventhubs.EventDataBatchOptions
}

const postEventHubName = "micce-search-engine"

func NewPostEventHubProducer(azureEventHubConnectionName string) (*EventHubProducer, error) {
	producerClient, err := azeventhubs.NewProducerClientFromConnectionString(azureEventHubConnectionName, postEventHubName, nil)
	option := azeventhubs.EventDataBatchOptions{
		//PartitionID: &partitionId,
	}
	if err != nil {
		return nil, err
	}
	return &EventHubProducer{
		producer: producerClient,
		option:   &option,
	}, nil

}

func (e *EventHubProducer) CreateEventBatch(ctx context.Context, eventDatas []azeventhubs.EventData) (*azeventhubs.EventDataBatch, error) {
	batchData, err := e.producer.NewEventDataBatch(ctx, e.option)
	if err != nil {
		return nil, err
	}
	for i := range eventDatas {
		err = batchData.AddEventData(&eventDatas[i], nil)
		if err != nil {
			return nil, err
		}
	}

	return batchData, nil
}

func (e *EventHubProducer) Send(ctx context.Context, batch *azeventhubs.EventDataBatch) error {
	err := e.producer.SendEventDataBatch(ctx, batch, nil)
	if err != nil {
		return err
	}
	return nil
}

func (e *EventHubProducer) Close(ctx context.Context) {
	err := e.producer.Close(ctx)
	if err != nil {
		internal.Logger.Error("fatal close event hub producer: %s", err.Error())
	}
}
