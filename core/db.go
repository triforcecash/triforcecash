package core

import (
	
	"log"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/util"
)

var LvlDB, _ = leveldb.OpenFile("data", nil)

func DBOpen() {
	LvlDB, _ = leveldb.OpenFile("data", nil)
}

func DBClose() {
	LvlDB.Close()
}

func Get(prfx string, key []byte) []byte {

	key1 := append([]byte(prfx), key...)
	data, err := LvlDB.Get(key1, nil)
	if err != nil {
		return nil
	}
	return data
}

func Put(prfx string, key, data []byte) {
	key1 := append([]byte(prfx), key...)
	LvlDB.Put(key1, data, nil)
}


func RemoveTmp(){
	iter:=LvlDB.NewIterator(util.BytesPrefix([]byte("tmp-")),nil)
	var c int
	for iter.Next(){
		LvlDB.Delete(iter.Key(),nil)
		c++
	}

	log.Printf("Removed %d values",c)
	iter.Release()
}