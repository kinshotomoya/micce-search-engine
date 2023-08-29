## indexer
search-engineにfirestoreのspotコレクションを全件upsertするシステム


### 仕様
- 単発バッチ
  - 検索エンジンをGCPサーバにデプロイした後、手動でこのバッチを動かすことで検索エンジンにデータを入れる
- 定期バッチ
  - 1日1回、検索エンジンのデータとfirestoreのデータの差分更新を行う
  - firestoreにあって、検索エンジンにないspotデータをinsertする
  - firestoreのデータが更新されていた場合、検索エンジンのデータを更新する


## 技術スタック
Java: 17
maven: 3.9.4



## ローカル開発

```shell
$ mvn package
$ java -jar target/micce-indexer-0.1.0.jar
```
