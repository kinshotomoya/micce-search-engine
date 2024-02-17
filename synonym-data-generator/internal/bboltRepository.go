package repository

import (
	"bufio"
	"go.etcd.io/bbolt"
	"os"
	"strings"
	"syscall"
	"time"
)

type BboltRepository struct {
	db         *bbolt.DB
	bucketName string
}

// NOTE: 所有者だけread write権限
const PERMISSION = 0600

func NewBboltRepositpry() (*BboltRepository, error) {
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
	db, err := bbolt.Open("synonym.db", PERMISSION, &option)
	if err != nil {
		return nil, err
	}
	return &BboltRepository{
		db:         db,
		bucketName: "synonymBucket",
	}, err
}

func (r *BboltRepository) Write(filePath string) error {

	open, err := os.Open(filePath)
	defer open.Close()
	if err != nil {
		return err
	}

	err = r.db.Batch(func(tx *bbolt.Tx) error {
		bucket, err := tx.CreateBucket([]byte(r.bucketName))
		if err != nil {
			return err
		}
		scanner := bufio.NewScanner(open)
		for scanner.Scan() {
			raw := scanner.Text()
			raws := strings.Split(raw, ",")
			err = bucket.Put([]byte(raws[0]), []byte(raws[1]))
			if err != nil {
				return err
			}
			if err != nil {
				return err
			}
		}
		return nil
	})

	if err != nil {
		return err
	}

	return nil
}
