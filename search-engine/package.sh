#!/bin/sh

ENV=$1
if [ -d "target" ]; then
	echo "デプロイパッケージが存在します。削除します。"
	rm -r target/
        echo "削除完了しました。"
fi

echo "デプロイパッケージが存在しません。作成します。"
mkdir target

cp -r components target/
cp -r schemas target/
cp -r $ENV/services.xml target/
cp -r $ENV/hosts.xml target/

echo "デプロイパッケージの作成が完了しました。"
