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
	"reader/internal/azure"
	firestore2 "reader/internal/firestore"
	time2 "reader/internal/time"
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

	// TODO: azureのcontainerで動かす場合は、環境変数にEVT_HUB_CONNECTION_NAMEを登録する
	azureEventHubConnectionName := os.Getenv("EVT_HUB_CONNECTION_NAME")

	fmt.Println(azureEventHubConnectionName)

	fireStoreClient, err := firestore2.NewClient(ctx)
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

	// NOTE: firestoreに保存したSpotScheduledTimeコレクションから開始日を取得する
	//  開始日から現在日時の間の更新データを全てqueueにつむ
	//  indexer側でvespaにフィードする流量を調整する
	doc, err := fireStoreClient.GetDocumentOne(ctx)
	run(ctx, fireStoreClient, azureEventHubProducer, &doc.Datetime)
	log.Println("finished scheduledTime upsert")

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
				//// TODO: ↓デバッグのために一年前のtimeを取得しているので1時間前に変更
				beforeOneHour := timeConf.BeforeOneYear()
				run(ctx, fireStoreClient, azureEventHubProducer, &beforeOneHour)
			}
		}
	}()

	ctxNotify, cancel := signal.NotifyContext(ctx, syscall.SIGTERM, syscall.SIGINT, syscall.SIGHUP)
	defer cancel()

	<-ctxNotify.Done()

	err = setScheduledTime(ctx, fireStoreClient, timeConf)

	if err != nil {
		log.Printf("error upsert scheduled time: %s", err.Error())
	}

	tickerDoneChanel <- true

	wg.Wait()
	log.Println("program exit")

}

func run(ctx context.Context, fireStoreClient *firestore2.FireStoreClient, azureEventHubProducer *azure.EventHubProducer, startTime *time.Time) {
	const EVENT_HUB_BATCH_SIZE = 100
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	documentIter := fireStoreClient.GetDocumentsByUpdateAt(ctx, *startTime)

	// 3並列でeventHubに送る
	ch := make(chan struct{}, 3)
	eventDatas := make([]azeventhubs.EventData, 0, EVENT_HUB_BATCH_SIZE)
	for {
		snapShot, err := documentIter.Next()
		if err != nil {
			log.Println("もうiteratorにないのでloop抜ける")
			break
		}
		doc := firestore2.CreateDocument(snapShot)
		var buf bytes.Buffer
		err = json.NewEncoder(&buf).Encode(doc)
		if err != nil {
			log.Printf("fatal encode firestore data to json: %s", err.Error())
		}
		eventData := azeventhubs.EventData{
			Body: buf.Bytes(),
		}
		eventDatas = append(eventDatas, eventData)
		if len(eventDatas) >= EVENT_HUB_BATCH_SIZE {
			// bufferを満たす
			// 3並列までの制御のため
			// 4つめgoroutineを作成する際にここで待ちが発生する
			ch <- struct{}{}
			goroutineEventDatas := make([]azeventhubs.EventData, 0, 100)
			// 別のgoroutine内でだけで利用するスライスなのでコピーする
			copy(goroutineEventDatas, eventDatas)
			// 元々のスライスを初期化(len=0, cap=100になる)
			eventDatas = eventDatas[:0]
			go func() {
				sendToEventHub(ctx, azureEventHubProducer, goroutineEventDatas)
				// NOTE: channel内のbufferを１つ解放
				<-ch
			}()
		}
	}
	if len(eventDatas) > 0 {
		sendToEventHub(ctx, azureEventHubProducer, eventDatas)
	}

}

func sendToEventHub(ctx context.Context, azureEventHubProducer *azure.EventHubProducer, eventDatas []azeventhubs.EventData) {
	timeoutCtx, cancel := context.WithTimeout(ctx, 1000*time.Millisecond)
	defer cancel()
	eventDataBatch, err := azureEventHubProducer.CreateEventBatch(timeoutCtx, eventDatas)
	if err != nil {
		log.Printf("fatal create eventhub data: %s", err.Error())
	}
	err = azureEventHubProducer.Send(timeoutCtx, eventDataBatch)
	if err != nil {
		log.Printf("fatal send eventhub data: %s", err.Error())
	}

}

func setScheduledTime(ctx context.Context, fireStoreClient *firestore2.FireStoreClient, timeConf *time2.Time) error {
	scheduledTime := timeConf.Now()
	data := make(map[string]time.Time)
	data["datetime"] = scheduledTime
	err := fireStoreClient.UpsertDocument(ctx, data)
	if err != nil {
		return err
	}
	return nil
}
