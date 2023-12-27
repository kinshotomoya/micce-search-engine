
ローカル動作確認
```shell
go run cmd/main.go -env dev
```

## 検索エンジン処理
 
◻️TODO:ブックマークしているスポット検索

・ブックマークの作成・更新・削除
ブックマークのデータは物理削除になってるのでreaderからは検知できない


◻️データベースにスポットが存在しない場合
これは今まで通りアプリサーバが処理行う（google APIからspot情報を取得して、pre eventHubにeventを送る）
=> 検索エンジンにindexするフローとしては、reader -> eventhubs -> indexerのフローの一系統にまとめたいので

