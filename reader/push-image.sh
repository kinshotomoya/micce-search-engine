#!/bin/zsh

docker build . -t reader
docker tag reader miccereaderacr.azurecr.io/reader
docker push miccereaderacr.azurecr.io/reader
