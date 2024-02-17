package repository

import (
	"fmt"
	"go.etcd.io/bbolt"
	"os"
	"syscall"
	"time"
)

type BboltRepository struct {
	db         *bbolt.DB
	bucketName string
}

// NOTE: 所有者だけread write権限
const PERMISSION = 0600
const DB_NAME = "synonym.db"
const BUCKET_NAME = "synonymBucket"

func NewBboltRepositpry() (*BboltRepository, error) {

	// まずDBの存在確認を行う
	_, err := os.Stat(DB_NAME)
	if err != nil {
		return nil, err
	}

	// NOTE: 結局内部的にはmmapを使っている
	option := bbolt.Options{
		Timeout: 1 * time.Second,
		// NOTE: マップされたメモリの修正がプロセス固有であると設定
		MmapFlags: syscall.MAP_PRIVATE,
		// NOTE: mmpaの初期サイズ（bytes）
		// 100MBを設定
		InitialMmapSize: 100 * 1024 * 1024,
		// NOTE: mmapでメモリに展開した領域をロックする（ページアウトさせない）
		// ページフォルトが発生しないので高速化するが、ユースケースがわからない
		// そもそもメモリに乗り切らないデータを扱うためにmmapを利用すると思うが、なぜmlockするのか
		Mlock: false,
		// NOTE: mmapでの割り当てはページ単位であり、その割り当てでフリーになっているつまり、どのデータも割り当てられていない
		// ページをリストを管理する方法を設定
		// 二種類あり、arrayとhashMap
		//  hashMapの方が高速
		FreelistType: bbolt.FreelistMapType,
	}
	db, err := bbolt.Open(DB_NAME, PERMISSION, &option)
	if err != nil {
		return nil, err
	}
	return &BboltRepository{
		db:         db,
		bucketName: BUCKET_NAME,
	}, err
}

func (r *BboltRepository) GetValue(key []byte) []byte {
	var value []byte
	r.db.View(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket([]byte(r.bucketName))
		value = bucket.Get(key)
		return nil
	})

	fmt.Println(string(value))
	return value
}
