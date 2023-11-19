#!/bin/zsh

# acrにログイン
# 参考：https://learn.microsoft.com/ja-jp/azure/container-registry/container-registry-authentication?tabs=azure-cli
az acr login -n micceindexeracr
docker build . -t indexer
docker tag indexer micceindexeracr.azurecr.io/indexer
docker push micceindexeracr.azurecr.io/indexer
