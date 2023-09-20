package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/Azure/azure-sdk-for-go/sdk/messaging/azeventhubs"
	"github.com/joho/godotenv"
	"log"
	"os"
	"os/signal"
	"reader/azure"
	"reader/firestore"
	time2 "reader/time"
	"sync"
	"syscall"
	"time"
)

func main() {
	ctx := context.Background()

	env := flag.String("env", "", "hoge")
	flag.Parse()

	if *env == "dev" {
		envErr := godotenv.Load()
		if envErr != nil {
			log.Fatal("error loading .env file")
		}
	}

	azureEventHubConnectionName := os.Getenv("EVT_HUB_CONNECTION_NAME")

	fmt.Println(azureEventHubConnectionName)

	fireStoreClient, err := firestore.NewClient(ctx)
	defer fireStoreClient.Close()
	if err != nil {
		log.Printf("Failed to create firestore client: %v", err)
	}

	azureEventHubProducer, err := azure.NewEventHubProducer(azureEventHubConnectionName)
	if err != nil {
		log.Fatalf("fatal create event hub producer: %s", err.Error())
	}
	defer azureEventHubProducer.Close(ctx)

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
				run(ctx, fireStoreClient, azureEventHubProducer, timeConf)
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

func run(ctx context.Context, fireStoreClient *firestore.FireStoreClient, azureEventHubProducer *azure.EventHubProducer, timeConf *time2.Time) {

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
		log.Println(snapShot.Data())
		doc := firestore.CreateDocument(snapShot)
		sendToEventHub(ctx, azureEventHubProducer, doc)
	}
}

func sendToEventHub(ctx context.Context, azureEventHubProducer *azure.EventHubProducer, data firestore.Document) {
	var buf bytes.Buffer
	err := json.NewEncoder(&buf).Encode(data)
	if err != nil {
		log.Fatalf("fatal encode firestore data to json: %s", err.Error())
	}

	eventData := &azeventhubs.EventData{
		Body: buf.Bytes(),
	}
	timeoutCtx, cancel := context.WithTimeout(ctx, 1000*time.Millisecond)
	defer cancel()
	eventDataBatch, err := azureEventHubProducer.CreateEventBatch(timeoutCtx, eventData)
	if err != nil {
		log.Printf("fatal create eventhub data: %s", err.Error())
	}
	err = azureEventHubProducer.Send(timeoutCtx, eventDataBatch)
	if err != nil {
		log.Printf("fatal send eventhub data: %s", err.Error())
	}

}
