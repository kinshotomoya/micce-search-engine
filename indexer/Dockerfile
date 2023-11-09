# syntax=docker/dockerfile:1
FROM golang:1.21-alpine AS builder

WORKDIR /app

COPY go.mod ./
COPY go.sum ./

RUN go mod download

COPY ./ ./
# -o: バイナリファイルを生成する場所を指定
# 参考：https://qiita.com/kanuma1984/items/80b1f4c35ba847bb84a2
# ldflags: go toolのシンボリックテーブルが除かれる
RUN go build -o ./bin/indexer -ldflags="-s -w" cmd/main.go

FROM debian:bullseye-slim

# ルート証明書をインストール
# debian-slimにはデフォルトで入っていない
# これがないとdocker内部からgo http clientを使って別apiをhttpリクエストする際に、以下エラーになる
# tls: failed to verify certificate: x509: certificate signed by unknown authority
RUN apt update && apt install -y ca-certificates

RUN groupadd -r indexer && useradd -r -g indexer indexer

USER indexer

WORKDIR /bin

COPY --from=builder /app/bin/indexer ./

CMD [ "/bin/indexer" ]