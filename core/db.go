package core

import (
	"github.com/syndtr/goleveldb/leveldb"
)

var LvlDB, _ = leveldb.OpenFile("data", nil)

func DBOpen() {
	LvlDB, _ = leveldb.OpenFile("data", nil)
}

func DBClose() {
	LvlDB.Close()
}

func Get(prfx, key []byte) []byte {

	key1 := append(prfx, key...)
	data, err := LvlDB.Get(key1, nil)
	if err != nil {
		return nil
	}
	return data
}

func Put(prfx, key, data []byte) {
	key1 := append(prfx, key...)
	LvlDB.Put(key1, data, nil)
}
