package main

import (
	"context"
	"indexer/gcp"
	"log"
	"os/signal"
	"syscall"
)

func main() {

	// 1.pubsubからmessageをstreamで取得
	// 2.vespaにドキュメントをupsert

	// NOTE: ↓streaming apiを使えば、subscriber側でポーリングせずにstreamでmessageを取得できる
	// https://christina04.hatenablog.com/entry/cloud-pubsub

	ctx := context.Background()
	pubsubClient, err := gcp.NewPubSubClient(ctx)

	if err != nil {
		log.Fatalf("fatal create pubsub pubsubClient: %s", err.Error())
	}

	ctx, cancel := signal.NotifyContext(ctx, syscall.SIGTERM, syscall.SIGINT, syscall.SIGHUP)
	defer cancel()

	log.Println("program is running")

	// cancel付きのcontextを生成
	// 終了シグナルを受け取った時点でcancelPubSub()を実行し、pubsubClient.Subscribe(canCtx)内のcanCtxにdoneチャネルを閉じる
	// Subscribe内のsub.Receive(ctx)内で、Doneチャネルが閉じられるとpubsub pullをやめる実装がされている
	canCtx, cancelPubSub := context.WithCancel(ctx)
	err = pubsubClient.Subscribe(canCtx)

	if err != nil {
		log.Fatalf("fatal subscribe pubsub: %s", err.Error())
	}

	<-ctx.Done()

	// 終了処理
	cancelPubSub()
	err = pubsubClient.Close()
	if err != nil {
		log.Fatalf("fatal close pubsub pubsubClient: %s", err.Error())
	}

	log.Println("program exit")

}
