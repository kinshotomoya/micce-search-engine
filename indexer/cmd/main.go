package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/joho/godotenv"
	"indexer/internal/azure"
	vespa2 "indexer/internal/vespa"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

func main() {

	// 1.eventhubからmessageをstreamで取得
	// 2.vespaにドキュメントをupsert

	// NOTE: ↓streaming apiを使えば、subscriber側でポーリングせずにstreamでmessageを取得できる
	// https://christina04.hatenablog.com/entry/cloud-pubsub

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
	azureStorageAccountConnectionName := os.Getenv("AZURE_STORAGE_ACCOUNT_CONNECTION_NAME")
	vespaHostName := os.Getenv("VespaHostName")

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

	config := &vespa2.VespaConfig{
		Url:     fmt.Sprintf("http://%s:8080", vespaHostName),
		Timeout: 1000,
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

	vespaClient := &vespa2.VespaClient{
		Client: httpClient,
		Config: config,
	}

	ctx, cancel := signal.NotifyContext(ctx, syscall.SIGTERM, syscall.SIGINT, syscall.SIGHUP)
	defer cancel()

	log.Println("program is running")

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		azure.DispatchPartitionClients(processor, ctx, vespaClient)
		log.Println("親goroutine終了")
	}()

	// processorを起動する
	log.Println("start processor")
	if err := processor.Run(ctx); err != nil {
		log.Fatalf("fatal run azure eventhub processor: %s", err.Error())
	}

	wg.Wait()

	log.Println("receive kill signal")

	vespaClient.Close()

	log.Println("program exit")

}
