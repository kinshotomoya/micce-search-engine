package firestore

import (
	"cloud.google.com/go/firestore"
	"context"
	"time"
)

type FireStoreClient struct {
	Client *firestore.Client
}

const projectID = "micce-travel"

func NewClient(ctx context.Context) (*FireStoreClient, error) {
	client, err := firestore.NewClient(ctx, projectID)
	if err != nil {
		return nil, err
	}
	return &FireStoreClient{
		Client: client,
	}, nil
}

func (client *FireStoreClient) Close() {
	client.Client.Close()
}

func (client *FireStoreClient) GetDocumentsByUpdateAt(ctx context.Context, now time.Time) *firestore.DocumentIterator {
	return client.Client.Collection("Spot").Where("updatedAt", ">=", now).Documents(ctx)
}
