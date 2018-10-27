package core

import (
	"bytes"
	"math/big"
	"sync"
)

const (
	KeysLimit = 10000
)

type Key struct {
	Pub   []byte
	Nonce []byte
	Rate  *big.Int
	Part  float64
}

func (self *Key) CalcRate() *big.Int {
	var buf bytes.Buffer
	buf.Write(self.Pub)
	buf.Write(self.Nonce)
	self.Rate = coef(buf.Bytes())
	return self.Rate
}

type KeysPool struct {
	Keys       map[string]*Key
	Mux        sync.Mutex
	Difficulty *big.Int
	Total      *big.Int
}

func (self *KeysPool) IncreaseDifficulty() {
	self.Difficulty.Mul(self.Difficulty, big.NewInt(101))
	self.Difficulty.Div(self.Difficulty, big.NewInt(100))
}

func (self *KeysPool) DecreaseDifficulty() {
	self.Difficulty.Mul(self.Difficulty, big.NewInt(99))
	self.Difficulty.Div(self.Difficulty, big.NewInt(100))
	if self.Difficulty.Cmp(big.NewInt(1000)) == -1 {
		self.Difficulty.SetInt64(1000)
	}
}

func (self *Key) Encode() []byte {
	return Join([][]byte{
		self.Pub,
		self.Nonce,
	})
}
func DecodeKey(blob []byte) *Key {
	Args := Listblob(Split(blob))
	return &Key{
		Pub:   Args.Get(0),
		Nonce: Args.Get(1),
	}
}

func (self *KeysPool) Encode() []byte {
	var res [][]byte

	self.Mux.Lock()
	defer self.Mux.Unlock()

	for _, k := range self.Keys {
		res = append(res, k.Encode())
	}
	return Join(res)
}

func (self *KeysPool) Add(key *Key) {

	if len(key.Encode()) > 250 {
		return
	}

	if key.CalcRate().Cmp(self.Difficulty) == 1 {

		self.Mux.Lock()
		defer self.Mux.Unlock()
		if key0, ok := self.Keys[string(key.Pub)]; ok {
			if key.Rate.Cmp(key0.Rate) == 1 {
				self.Total.Sub(self.Total, key0.Rate)
				self.Total.Add(self.Total, key.Rate)
				self.Keys[string(key.Pub)] = key
			}
		} else {
			self.Total.Add(self.Total, key.Rate)
			self.Keys[string(key.Pub)] = key
		}
	}
}

func (self *KeysPool) AddNew(blob []byte) {
	l := Split(blob)
	for _, k := range l {
		self.Add(DecodeKey(k))
	}
}

func (self *KeysPool) Sync() {
	res := Peers.Request(
		Join(
			[][]byte{
				[]byte("sync keys"),
				[]byte{},
			}),
		func(blob []byte) bool {
			return true
		},
	)
	self.AddNew(res)

	if len(self.Keys) > KeysLimit {
		self.IncreaseDifficulty()
	} else {
		self.DecreaseDifficulty()
	}

	self.DeleteLowRate()
}

func (self *KeysPool) DeleteLowRate() {
	self.Mux.Lock()
	defer self.Mux.Unlock()
	for k, key := range self.Keys {
		if key.Rate.Cmp(self.Difficulty) == -1 {
			self.Total.Sub(self.Total, key.Rate)
			delete(self.Keys, k)
		}
	}
}

func (self *KeysPool) Map(fun func(key *Key)) {
	self.Mux.Lock()
	for _, key := range self.Keys {
		self.Mux.Unlock()
		fun(key)
		self.Mux.Lock()
	}
	self.Mux.Unlock()
}

func (self *KeysPool) AddSelf() {
	self.Add(&Key{
		Pub:   Pub,
		Nonce: Nonce,
	})
}
