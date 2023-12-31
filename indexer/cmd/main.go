package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/joho/godotenv"
	"indexer/internal"
	"indexer/internal/domain/model"
	"indexer/internal/repository/azure"
	"indexer/internal/repository/mysql"
	"indexer/internal/repository/vespa"
	"indexer/internal/service"
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
	// TODO: 3. rdbにvespa更新完了記録する

	internal.InitLogger()

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

	consumerProcessor, err := azure.NewPreEventHubConsumerClient(azureEventHubConnectionName, azureStorageAccountConnectionName)
	if err != nil {
		internal.Logger.Error(fmt.Sprintf("fatal create pre event hub consumer: %s", err.Error()))
	}

	config := &vespa.VespaConfig{
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

	vespaClient := &vespa.VespaClient{
		Client: httpClient,
		Config: config,
	}

	repository, err := mysql.NewMysqlRepository()
	if err != nil {
		internal.Logger.Error(fmt.Sprintf("Failed to create mysql repository: %v", err))
	}

	customTime, err := model.NewCustomTime()
	if err != nil {
		internal.Logger.Error(fmt.Sprintf("Failed to create custom time: %v", err))
	}

	readService := &service.ReadService{
		EventHubConsumerProcessor: consumerProcessor,
		MysqlRepository:           repository,
		VespaClient:               vespaClient,
		CustomTime:                customTime,
	}

	ctx, cancel := signal.NotifyContext(ctx, syscall.SIGTERM, syscall.SIGINT, syscall.SIGHUP)
	defer cancel()

	var wg sync.WaitGroup

	// processorを起動する
	wg.Add(1)
	go func() {
		defer wg.Done()
		err = consumerProcessor.Run(ctx)
		if err != nil {
			internal.Logger.Error(fmt.Sprintf("fatal run cunsumerProcessor: %s", err))
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		err = readService.Run(ctx)
		if err != nil {
			internal.Logger.Error(fmt.Sprintf("fatal run cunsumerProcessor: %s", err))
		}
	}()

	wg.Wait()

	internal.Logger.Info("receive kill signal")

	vespaClient.Close()

	internal.Logger.Info("program exit")

}
