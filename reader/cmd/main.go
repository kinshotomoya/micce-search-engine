package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/joho/godotenv"
	"os"
	"os/signal"
	"reader/internal"
	"reader/internal/domain/model"
	azure2 "reader/internal/repository/azure"
	"reader/internal/repository/firestore"
	"reader/internal/repository/mysql"
	"reader/internal/service"
	"sync"
	"syscall"
)

func main() {
	ctx := context.Background()

	internal.InitLogger()

	env := flag.String("env", "", "hoge")
	flag.Parse()

	if *env == "dev" {
		envErr := godotenv.Load()
		if envErr != nil {
			internal.Logger.Error("error loading .env file")
		}
	}

	azureEventHubConnectionName := os.Getenv("EVT_HUB_CONNECTION_NAME")
	azureStorageAccountConnectionName := os.Getenv("AZURE_STORAGE_ACCOUNT_CONNECTION_NAME")

	fireStoreClient, err := firestore.NewClient(ctx)
	defer fireStoreClient.Close()
	if err != nil {
		internal.Logger.Error(fmt.Sprintf("Failed to create firestore client: %v", err))
	}

	// pre-eventhubからイベントを取得するconsumerの作成
	consumerProcessor, err := azure2.NewPreEventHubConsumerClient(azureEventHubConnectionName, azureStorageAccountConnectionName)
	if err != nil {
		internal.Logger.Error(fmt.Sprintf("fatal create pre event hub consumer: %s", err.Error()))
	}

	// post-eventhubにイベントを送るproducerの作成
	azureEventHubProducer, err := azure2.NewPostEventHubProducer(azureEventHubConnectionName)
	if err != nil {
		internal.Logger.Error(fmt.Sprintf("fatal create event hub producer: %s", err.Error()))
	}
	defer azureEventHubProducer.Close(ctx)

	if err != nil {
		internal.Logger.Error(fmt.Sprintf("Failed to create post event hub producer: %v", err))
	}

	repository, err := mysql.NewMysqlRepository()
	if err != nil {
		internal.Logger.Error(fmt.Sprintf("Failed to create mysql repository: %v", err))
	}

	customTime, err := model.NewCustomTime()
	if err != nil {
		internal.Logger.Error(fmt.Sprintf("Failed to create custom time: %v", err))
	}

	firestoreClient, err := firestore.NewClient(ctx)
	if err != nil {
		internal.Logger.Error(fmt.Sprintf("Failed to create firestoreClient: %v", err))
	}

	readService := &service.ReadService{
		EventHubConsumerProcessor: consumerProcessor,
		EventHubProducer:          azureEventHubProducer,
		MysqlRepository:           repository,
		CustomTime:                customTime,
		FireStoreClient:           firestoreClient,
	}

	withCancel, readServiceCancelFunc := context.WithCancel(ctx)

	// consumerProcessorを起動
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		err = consumerProcessor.Run(withCancel)
		if err != nil {
			internal.Logger.Error(fmt.Sprintf("fatal run cunsumerProcessor: %s", err))
		}
		wg.Done()
	}()

	wg.Add(1)
	go func() {
		err = readService.Run(withCancel)
		if err != nil {
			internal.Logger.Error(fmt.Sprintf("error run eventhub Consumer: %s", err.Error()))
		}
		wg.Done()
	}()

	// 終了シグナル待ち受け
	ctxNotify, cancel := signal.NotifyContext(ctx, syscall.SIGTERM, syscall.SIGINT, syscall.SIGHUP)
	defer cancel()
	<-ctxNotify.Done()
	// SIGTERMを受け取ったらeventHubConsumerクライアントそれぞれ終了させる（chanel閉じる）
	internal.Logger.Info("signal received")
	readServiceCancelFunc()
	wg.Wait()
	internal.Logger.Info("program exit")
}
