package main

import (
	"bytes"
	"cloud.google.com/go/pubsub"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"reader/firestore"
	"reader/gcp"
	time2 "reader/time"
	"time"
)

func main() {
	ctx := context.Background()
	fireStoreClient, err := firestore.NewClient(ctx)
	defer fireStoreClient.Close()
	if err != nil {
		log.Printf("Failed to create firestore client: %v", err)
	}

	gcpClient, err := gcp.NewPubSubClient(ctx)
	if err != nil {
		log.Printf("Failed to create gcp client: %v", err)
	}

	timeConf := time2.NewTime()
	// TODO: ↓デバッグのために一年前のtimeを取得しているので1時間前に変更
	beforeOneHour := timeConf.BeforeOneYear()
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	documentIter := fireStoreClient.GetDocumentsByUpdateAt(ctx, beforeOneHour)

	for {
		snapShot, err := documentIter.Next()
		if err != nil {
			// もうないiterator中身がない場合ループを抜ける
			fmt.Println("もうiteratorにないのでloop抜ける")
			break
		}
		doc := firestore.CreateDocument(snapShot)
		publishToPubSub(ctx, gcpClient, doc)
	}

}

// NOTE: ↓pubsub clientを使ってpublishする
// https://pkg.go.dev/cloud.google.com/go/pubsub#section-readme

// ↓pubsubの解説
// https://laboratory.kiyono-co.jp/69/gcp/
// https://cloud.google.com/pubsub/docs/pull?hl=ja#go
// https://cloud.google.com/pubsub/docs/pull?hl=ja#java_3

func publishToPubSub(ctx context.Context, gcpClient *gcp.PubSubClient, data firestore.Document) {
	var buf bytes.Buffer
	err := json.NewEncoder(&buf).Encode(data)
	if err != nil {
		log.Fatalf("fatal encode firestore data to json: %s", err.Error())
	}
	message := &pubsub.Message{
		Data: buf.Bytes(),
	}
	gcpClient.Publish(ctx, message)
}
