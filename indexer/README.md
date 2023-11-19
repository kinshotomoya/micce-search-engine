

```shell
$ go run cmd/main.go -env dev
```

※ GPCのmicce-travel-appプロジェクトを利用する


## azureへのデプロイ方法

1. コンテナイメージをazureコンテナレジストリにアップロード
```shell
$ ./push-image.sh
```

2. azureポータルでコンテナインスタンスを再起動
最新のイメージでデプロイされる