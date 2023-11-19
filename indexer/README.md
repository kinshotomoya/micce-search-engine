## ローカル起動
```shell
$ go run cmd/main.go -env dev
```

## azureへのデプロイ方法

1. コンテナイメージをazureコンテナレジストリにアップロード
```shell
$ ./push-image.sh
```

2. azureポータルでコンテナインスタンスを再起動
最新のイメージでデプロイされる