package firestore

import (
	"cloud.google.com/go/firestore"
	"context"
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

func (client *FireStoreClient) GetDocumentsBySpotIds(ctx context.Context, spotIds []string) *firestore.DocumentIterator {
	return client.Client.Collection("Spot").Where("", "in", spotIds).Documents(ctx)
}
