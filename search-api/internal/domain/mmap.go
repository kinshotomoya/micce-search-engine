package domain

import (
	"fmt"
	"github.com/edsrzf/mmap-go"
	"os"
	"sync"
)

type CustomMMap struct {
	m    *sync.RWMutex
	mmap mmap.MMap
	// NOTE: offsetはos.Getpagesize()の倍数でないといけない
	offset       int64
	fullFileSize int64
	chunkSize    int
}

func NewCustomMMap() (*CustomMMap, error) {
	file, err := os.Open("dict/synonym.csv")
	if err != nil {
		return nil, err
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		return nil, err
	}

	// ファイルサイズ（単位: byte）
	fileSize := fileInfo.Size()
	// オフセット
	initOffSet := int64(0)

	// TODO 方針：
	//  1. synonym.csvから、↑csvをバイナリにした時のkeyとkeyの位置（offset）のインデックスを作成する（検索時にはmmapにマップしたバイナリデータに対して検索を行うから）
	//    別のツールでインデックスファイルを作成して、search-api起動時にメモリに配置する
	//    1.1 todoインデックスファイルも全部メモリではなくてファイルに保存しておいてmmapで取得するようにしたい
	//  2. synonym.csvをmmapでメモリにマップする
	//  3. 検索時には、検索文字列をバイナリにして、まずメモリにあるインデックス（hashmap）に対してoffsetを取得しにいく
	//  4. 取得したoffsetでmmapを検索するその時にoffset値から直近の[44]（カンマ）より左、[10]（改行）までのバイナリを取得する
	//    ex) [227 130 185 227 130 191 227 131 144 44 227 130 185 227 130 191 227 131 188 227 131 144 227 131 131 227 130 175 227 130 185 10 83 116 97 114 98]
	//    offset:0なら、[227 130 185 227 130 191 227 131 188 227 131 144 227 131 131 227 130 175 227 130 185]を取
	//    4.1 mmapにない場合、新規でファイルからmmapでメモリにファイル内容をマップする
	//  5. ↑で取得したバイナリを文字列に直す
	// : 探索する際にmmapに対象データがない場合は新規メモリマップし、offset値をみてそこからchunksize分ファイルの中身を更新する
	//  offset+chunksizeがfilesizeを超える場合は、超えた分はoffset:0から始める
	var newMmap mmap.MMap

	// 100MBをmmapでメモリにマップする
	// 1KB = 1024byte
	// 1MB = 1024KB
	// 100MB = 100 * 1MB
	chunkSize := 1024 * 1024 * 100
	if fileSize <= int64(chunkSize) {
		chunkSize = int(fileSize)
	}
	newMmap, err = mmap.MapRegion(file, chunkSize, mmap.RDONLY, 0, initOffSet)
	if err != nil {
		return nil, err
	}

	fmt.Println([]byte(","))
	// -> [44]
	fmt.Println([]byte("\n"))
	// -> [10]
	fmt.Println(string([]byte{227, 130, 185, 227, 130, 191, 227, 131, 144}))
	// -> スタバ
	fmt.Println(newMmap)

	return &CustomMMap{
		mmap:         newMmap,
		offset:       initOffSet,
		fullFileSize: fileSize,
		chunkSize:    chunkSize,
	}, nil
}
