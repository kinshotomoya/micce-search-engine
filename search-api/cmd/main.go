package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"github.com/joho/godotenv"
	"log/slog"
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
			panic("error loading .env file")
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

	bblot, err := repository.NewBboltRepositpry()
	if err != nil {
		slog.Error(err.Error())
		return
	}

	vespaRepository := repository.NewVespaRepository(
		model.NewVespaClient(httpClient, vespaUrl),
		bblot,
	)

	handler := presentation.NewHandler(vespaRepository)

	http.HandleFunc("/health", handler.HealthHandler)
	http.HandleFunc("/api/v1/search", handler.SearchHandler)

	baseCtx := context.Background()
	signalCtx, cancel := signal.NotifyContext(baseCtx, syscall.SIGTERM, syscall.SIGINT, syscall.SIGHUP)
	defer cancel()

	go func() {
		slog.Info("server is running")
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			panic(fmt.Sprintf("server is not running. %s", err))
		}
	}()

	<-signalCtx.Done()
	slog.Info("signal received")

	timeoutCtx, timeoutCancel := context.WithTimeout(baseCtx, 5*time.Second)
	defer timeoutCancel()
	err = server.Shutdown(timeoutCtx)
	if err != nil {
		slog.Info("fatal shutdown server. %s\n", err)
		return
	}
	vespaRepository.Close()

}
