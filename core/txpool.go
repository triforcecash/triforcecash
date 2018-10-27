package core

import (
	"sort"
	"sync"
)

const (
	MaxTxsNum = 2000
)

type TxsPool struct {
	Txs map[string]*Tx
	Mux sync.Mutex
}

func (self *TxsPool) Sync() {
	res := Peers.Request(Join([][]byte{
		[]byte("sync txspool"),
		[]byte{},
	}),
		func(blob []byte) bool {
			return true
		})
	self.AddNew(res)
}

func (self *TxsPool) Add(tx *Tx) {
	self.Mux.Lock()
	defer self.Mux.Unlock()
	if len(self.Txs) > MaxTxsNum {
		return
	}

	if tx.Check() {
		self.Txs[string(tx.SelfHash())] = tx
	}

}

func (self *TxsPool) Encode() []byte {
	self.Mux.Lock()
	defer self.Mux.Unlock()
	var res [][]byte
	for _, t := range self.Txs {
		res = append(res, t.Encode())
	}
	return Join(res)
}

func (self *TxsPool) AddNew(blob []byte) {
	res := Split(blob)
	for _, t := range res {
		self.Add(DecodeTx(t))
	}
}

func (self *TxsPool) ToList() TxsList {
	self.Mux.Lock()
	defer self.Mux.Unlock()
	txslist := TxsList{}
	for _, t := range self.Txs {
		txslist = append(txslist, t)
	}
	return txslist
}

func (self *TxsPool) GetTxs() TxsList {
	txslist := self.ToList()
	sort.Slice(txslist, func(i, j int) bool {
		a := txslist[i]
		b := txslist[j]
		return float64(a.Fee)/float64(len(a.Encode())) > float64(b.Fee)/float64(len(b.Encode()))
	})

	l := 0
	txs := TxsList{}
	for _, tx := range txslist {
		l += len(tx.Encode()) + 3
		if l > txsmaxlen {
			break
		}
		txs = append(txs, tx)
	}

	self.DeleteTxs(txs)

	return txs

}

func (self *TxsPool) DeleteTxs(txslist TxsList) {
	self.Mux.Lock()
	defer self.Mux.Unlock()

	for _, t := range txslist {
		delete(self.Txs, string(t.SelfHash()))
	}
	return
}
