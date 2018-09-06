package core

import "sort"

func SortPool() {
	poolmux.Lock()
	sort.Slice(TxsPool, func(i, j int) bool {
		a := TxsPool[i]
		b := TxsPool[j]
		return float64(a.Fee)/float64(len(a.Encode())) > float64(b.Fee)/float64(len(b.Encode()))
	})
	poolmux.Unlock()
}

func GetTxsFromPool() TxsList {
	SortPool()
	poolmux.Lock()
	l := 0
	txslist := TxsList{}
	for _, tx := range TxsPool {
		l += len(tx.Encode()) + 3
		if l > txsmaxlen {
			break
		}
		txslist = append(txslist, tx)
	}
	TxsPool = TxsPool[len(txslist):]
	poolmux.Unlock()
	return txslist
}
