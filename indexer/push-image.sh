#!/bin/zsh

docker build . -t indexer
docker tag indexer micceindexeracr.azurecr.io/indexer
docker push micceindexeracr.azurecr.io/indexer
