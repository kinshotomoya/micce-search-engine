#!/bin/zsh

# acrにログイン
# 参考：https://learn.microsoft.com/ja-jp/azure/container-registry/container-registry-authentication?tabs=azure-cli
az acr login -n miccereaderacr
docker build . -t reader
docker tag reader miccereaderacr.azurecr.io/reader
docker push miccereaderacr.azurecr.io/reader
