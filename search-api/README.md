## search-api
search-engineに検索をかける前段のAPI

## ローカル環境での動かし方
```shell
$ go run cmd/main.go -env dev
```

## 本番環境へのリクエスト方法

notionにhost名記載している
https://www.notion.so/micce/3c8982c225784ed0871a15a6b191f2d3?v=ab458fc1d1cc4a729380a26ae7296866

```shell
$ curl -i https://{hoge}/api/v1/search -d '{"spot_name": "京都", "limit": 10, "page": 1}'
```

## インターフェース
openapi.ymlを参照

## TODO: シノニム辞書
シノニム辞書をmmapで読み込んでクエリ発行時にOR検索させるようにする

## 今後
- internalなAPIなのでgRPCにしたい