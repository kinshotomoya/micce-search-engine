package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/joho/godotenv"
	"indexer/azure"
	"indexer/gcp"
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
	pubsubClient, err := gcp.NewPubSubClient(ctx)

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
		log.Fatalf("fatal create pubsub pubsubClient: %s", err.Error())
	}

	// TODO: 続きprocessorを作る箇所から
	//  https://github.com/Azure/azure-sdk-for-go/blob/main/sdk/messaging/azeventhubs/example_consuming_with_checkpoints_test.go#L54C1-L54C1

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

	// cancel付きのcontextを生成
	// 終了シグナルを受け取った時点でcancelPubSub()を実行し、pubsubClient.Subscribe(canCtx)内のcanCtxにdoneチャネルを閉じる
	// Subscribe内のsub.Receive(ctx)内で、Doneチャネルが閉じられるとpubsub pullをやめる実装がされている
	canCtx, cancelPubSub := context.WithCancel(ctx)
	err = pubsubClient.Subscribe(canCtx, vespaClient)

	if err != nil {
		log.Fatalf("fatal subscribe pubsub: %s", err.Error())
	}

	<-ctx.Done()

	// 終了処理
	cancelPubSub()
	err = pubsubClient.Close()
	vespaClient.Close()
	if err != nil {
		log.Fatalf("fatal close pubsub pubsubClient: %s", err.Error())
	}

	log.Println("program exit")

}
