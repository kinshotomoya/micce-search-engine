package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/joho/godotenv"
	"indexer/azure"
	"indexer/vespa"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {

	// 1.eventhubからmessageをstreamで取得
	// 2.vespaにドキュメントをupsert

	// NOTE: ↓streaming apiを使えば、subscriber側でポーリングせずにstreamでmessageを取得できる
	// https://christina04.hatenablog.com/entry/cloud-pubsub

	ctx := context.Background()
	//pubsubClient, err := gcp.NewPubSubClient(ctx)

	env := flag.String("env", "", "hoge")
	flag.Parse()

	if *env == "dev" {
		envErr := godotenv.Load()
		if envErr != nil {
			log.Fatal("error loading .env file")
		}
	}

	// TODO: azureのcontainerで動かす場合は、環境変数にEVT_HUB_CONNECTION_NAME,AZURE_STORAGE_ACCOUNT_CONNECTION_NAMEを登録する
	azureEventHubConnectionName := os.Getenv("EVT_HUB_CONNECTION_NAME")
	azureStorageAccountConnectionName := os.Getenv("AZURE_STORAGE_ACCOUNT_CONNECTION_NAME")

	fmt.Println(azureEventHubConnectionName, azureStorageAccountConnectionName)

	eventHubConsumer, checkpointStore, err := azure.NewEventHubConsumerClient(azureEventHubConnectionName, azureStorageAccountConnectionName)
	defer eventHubConsumer.Close(ctx)

	if err != nil {
		log.Fatalf("fatal create azure eventHubConsumer: %s", err.Error())
	}

	processor, err := azure.NewProcessor(eventHubConsumer, checkpointStore)

	if err != nil {
		log.Fatalf("fatal create azure eventhub processor: %s", err.Error())
	}

	config := &vespa.VespaConfig{
		Url:     "http://localhost:8080",
		Timeout: 1000,
	}

	// TODO: 検証・本番の場合は対象のvespa URLを設定する
	if *env != "dev" {
		config.Url = ""
	}

	transport := http.Transport{
		MaxIdleConns:       100,
		MaxConnsPerHost:    100,
		DisableKeepAlives:  false,
		IdleConnTimeout:    100 * time.Second,
		DisableCompression: true,
	}

	httpClient := &http.Client{
		Transport: &transport,
		// 参考：https://christina04.hatenablog.com/entry/go-timeouts
		Timeout: 90 * time.Second,
	}

	vespaClient := &vespa.VespaClient{
		Client: httpClient,
		Config: config,
	}

	ctx, cancel := signal.NotifyContext(ctx, syscall.SIGTERM, syscall.SIGINT, syscall.SIGHUP)
	defer cancel()

	log.Println("program is running")

	processCtx, processorCancel := context.WithCancel(ctx)
	ch := make(chan string)

	go azure.DispatchPartitionClients(processor, processCtx, ch)

	// processorを起動する
	log.Println("start processor")
	if err := processor.Run(processCtx); err != nil {
		log.Fatalf("fatal run azure eventhub processor: %s", err.Error())
	}

	<-ctx.Done()
	log.Println("receive kill signal")

	// 終了処理
	// partitionClientが動いている孫goroutineに終了シグナルを送る
	processorCancel()

	select {
	case msg := <-ch:
		log.Println(msg)
	case <-processCtx.Done():
		log.Println("no partition client closed cuz any partition client not running")
	}

	vespaClient.Close()

	log.Println("program exit")

}
