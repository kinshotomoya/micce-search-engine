#!/bin/zsh

# acrにログイン
# 参考：https://learn.microsoft.com/ja-jp/azure/container-registry/container-registry-authentication?tabs=azure-cli
az acr login -n miccesearchapiacr
docker build . -t search-api
docker tag search-api miccesearchapiacr.azurecr.io/searchapi
docker push miccesearchapiacr.azurecr.io/searchapi
