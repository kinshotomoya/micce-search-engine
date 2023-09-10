package gcp

import (
	"cloud.google.com/go/pubsub"
	"context"
	"fmt"
	"time"
)

type PubSubClient struct {
	Client *pubsub.Client
}

const projectID = "micce-travel"
const topicID = "micce-search-engine"

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

func (p *PubSubClient) Publish(ctx context.Context, message *pubsub.Message) {
	gcpCtx, gcpCancel := context.WithTimeout(ctx, 10*time.Second)

	// NOTE: publishする度にtopic作成して、publish後にバックグラウンドのgoroutineを停止させる
	topic := p.Client.Topic(topicID)
	result := topic.Publish(gcpCtx, message)
	defer topic.Stop()
	id, err := result.Get(ctx)
	if err != nil {
		fmt.Println(err.Error())
	} else {
		fmt.Printf("publish success: %s", id)
	}
	defer gcpCancel()
}
