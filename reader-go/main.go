package main

import (
	"context"
	"fmt"
	"log"
	"reader/firestore"
	time2 "reader/time"
	"time"
)

func main() {
	ctx := context.Background()
	fireStoreClient, err := firestore.NewClient(ctx)
	defer fireStoreClient.Close()
	if err != nil {
		log.Printf("Failed to create client: %v", err)
	}

	timeConf := time2.NewTime()
	// TODO: ↓デバッグのために一年前のtimeを取得しているので1時間前に変更
	beforeOneHour := timeConf.BeforeOneYear()
	fmt.Println(beforeOneHour)
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	documentIter := fireStoreClient.GetDocumentsByUpdateAt(ctx, beforeOneHour)

	// NOTE: メモリ節約のため、iteratorで20件づつpubsub(queue)に詰めるようにする
	//  容量20のsliceを作成
	docSlice := make([]firestore.Document, 0, 20)
	i := 0
	for {
		snapShot, err := documentIter.Next()
		if err != nil {
			// もうないiterator中身がない場合ループを抜ける
			log.Printf("pubsubにpublishする")
			break
		}

		doc := firestore.CreateDocument(snapShot)
		docSlice = append(docSlice, doc)

		if len(docSlice) == 20 {
			fmt.Println(docSlice)
			log.Printf("pubsubにpublishする")
			// NOTE: sliceを初期する（アロケートされているメモリはそのままで）
			docSlice = docSlice[:0]
		}

		i += i + 1

	}

}
