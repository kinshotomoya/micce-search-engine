package main

import (
	"context"
	"errors"
	"flag"
	"github.com/joho/godotenv"
	"log"
	"net/http"
	"os"
	"os/signal"
	"search-api/internal/presentation"
	"search-api/internal/repository"
	"search-api/internal/repository/model"
	"syscall"
	"time"
)

func main() {
	server := &http.Server{
		Addr: ":8081",
	}

	env := flag.String("env", "", "環境変数取得")
	flag.Parse()

	if *env == "dev" {
		envErr := godotenv.Load()
		if envErr != nil {
			log.Fatal("error loading .env file")
		}
	}

	vespaUrl := os.Getenv("VESPA_URL")

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

	vespaRepository := repository.NewVespaRepository(model.NewVespaClient(httpClient, vespaUrl))

	handler := presentation.NewHandler(vespaRepository)

	http.HandleFunc("/api/v1/search", handler.SearchHandler)

	baseCtx := context.Background()
	signalCtx, cancel := signal.NotifyContext(baseCtx, syscall.SIGTERM, syscall.SIGINT, syscall.SIGHUP)
	defer cancel()

	go func() {
		log.Println("server is running")
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("server is not running. %s", err)
		}
	}()

	<-signalCtx.Done()

	timeoutCtx, timeoutCancel := context.WithTimeout(baseCtx, 5*time.Second)
	defer timeoutCancel()
	err := server.Shutdown(timeoutCtx)
	if err != nil {
		log.Printf("fatal shutdown server. %s\n", err)
		return
	}

}
