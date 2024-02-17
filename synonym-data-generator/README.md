## 概要


## シノニム辞書追加方法

- dict/synonym.csvにcsv形式でシノニム展開したいspotNameを追加

ex)
スタバ,スターバックス
の場合は、ユーザー入力が「スタバ」の場合に「スターバック」とシノニム展開される


- 以下のコマンド打ってsynonym.dbを作成

```shell
 $ go run cmd/main.go
```

- synonym.dbが作成されたら、search-api/に持っていく
search-apiでは、search-api/配下のsynonym.dbを参照している