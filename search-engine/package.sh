#!/bin/sh

ENV=$1
if [ -d "target" ]; then
	echo "target ディレクトリが存在します。削除します。"
	rm -r target/
fi

echo "target ディレクトリが存在しません。作成します。"
mkdir target

cp -r components target/
cp -r schemas target/
cp -r $ENV target/

