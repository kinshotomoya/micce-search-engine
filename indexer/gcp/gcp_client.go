package gcp

import (
	"cloud.google.com/go/pubsub"
	"context"
	"fmt"
)

type PubSubClient struct {
	Client *pubsub.Client
}

const projectID = "micce-travel"
const subscriptionId = "micce-search-engine-sub"

func NewPubSubClient(ctx context.Context) (*PubSubClient, error) {
	c, err := pubsub.NewClient(ctx, projectID)
	if err != nil {
		fmt.Printf("fatal create pubsub client: %s", err.Error())
		return nil, err
	}
	return &PubSubClient{
		Client: c,
	}, nil
}

func (p *PubSubClient) Subscribe(ctx context.Context) error {
	sub := p.Client.Subscription(subscriptionId)
	err := sub.Receive(ctx, func(ctx context.Context, message *pubsub.Message) {
		fmt.Println(message.Data)
		message.Ack()
	})

	if err != nil {
		return err
	}

	return nil

}

func (p *PubSubClient) Close() error {
	err := p.Client.Close()
	if err != nil {
		return err
	}
	return nil
}