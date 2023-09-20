package azure

import (
	"context"
	"github.com/Azure/azure-sdk-for-go/sdk/messaging/azeventhubs"
	"log"
)

type EventHubProducer struct {
	producer *azeventhubs.ProducerClient
	option   *azeventhubs.EventDataBatchOptions
}

const eventHubName = "micce-search-engine"

func NewEventHubProducer(azureEventHubConnectionName string) (*EventHubProducer, error) {
	producerClient, err := azeventhubs.NewProducerClientFromConnectionString(azureEventHubConnectionName, eventHubName, nil)
	option := azeventhubs.EventDataBatchOptions{}
	if err != nil {
		return nil, err
	}
	return &EventHubProducer{
		producer: producerClient,
		option:   &option,
	}, nil

}

func (e EventHubProducer) CreateEventBatch(ctx context.Context, eventData *azeventhubs.EventData) (*azeventhubs.EventDataBatch, error) {
	batchData, err := e.producer.NewEventDataBatch(ctx, e.option)
	if err != nil {
		return nil, err
	}
	err = batchData.AddEventData(eventData, nil)
	if err != nil {
		return nil, err
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
		log.Fatalf("fatal close event hub producer: %s", err.Error())
	}
}
