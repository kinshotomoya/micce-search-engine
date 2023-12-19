package firestore

import (
	"cloud.google.com/go/firestore"
	"context"
	"log"
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

func (client *FireStoreClient) UpsertDocument(ctx context.Context, data any) error {
	_, err := client.Client.Collection("SpotScheduledTime").Doc("micce-search-engine").Set(ctx, data)
	if err != nil {
		return err
	}
	log.Print("set scheduled time")
	return nil
}

func (client *FireStoreClient) GetDocumentOne(ctx context.Context) (*SpotScheduledTime, error) {
	doc, err := client.Client.Collection("SpotScheduledTime").Doc("micce-search-engine").Get(ctx)
	if err != nil {
		return nil, err
	}
	var s SpotScheduledTime
	err = doc.DataTo(&s)
	if err != nil {
		return nil, err
	}
	return &s, nil
}
