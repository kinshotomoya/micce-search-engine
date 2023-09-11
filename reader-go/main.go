package main

import (
	"bytes"
	"cloud.google.com/go/pubsub"
	"context"
	"encoding/json"
	"log"
	"os/signal"
	"reader/firestore"
	"reader/gcp"
	time2 "reader/time"
	"sync"
	"syscall"
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
	defer gcpClient.Client.Close()

	if err != nil {
		log.Printf("Failed to create gcp client: %v", err)
	}

	timeConf := time2.NewTime()

	// TODO: 1時間ごとに変更
	ticker := time.NewTicker(30 * time.Second)
	tickerDoneChanel := make(chan bool)

	// 別スレッドでticker実行
	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		for {
			select {
			case <-tickerDoneChanel:
				log.Println("ticker stop")
				wg.Done()
				break
			case <-ticker.C:
				log.Println("running...")
				run(ctx, fireStoreClient, gcpClient, timeConf)
			}
		}

	}()

	ctx, cancel := signal.NotifyContext(ctx, syscall.SIGTERM, syscall.SIGINT, syscall.SIGHUP)
	defer cancel()

	<-ctx.Done()

	tickerDoneChanel <- true

	wg.Wait()
	log.Println("program exit")

}

func run(ctx context.Context, fireStoreClient *firestore.FireStoreClient, gcpClient *gcp.PubSubClient, timeConf *time2.Time) {

	//// TODO: ↓デバッグのために一年前のtimeを取得しているので1時間前に変更
	beforeOneHour := timeConf.BeforeOneYear()
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	documentIter := fireStoreClient.GetDocumentsByUpdateAt(ctx, beforeOneHour)

	for {
		snapShot, err := documentIter.Next()
		if err != nil {
			// もうないiterator中身がない場合ループを抜ける
			log.Println("もうiteratorにないのでloop抜ける")
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

// NOTE: ↓streaming apiを使えば、subscriber側でポーリングせずにstreamでmessageを取得できる
// https://christina04.hatenablog.com/entry/cloud-pubsub
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
