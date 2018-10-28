package core

import (
	"bytes"
)

func (self *Header) Cache(signs, block bool, checktime int64) []byte {
	d := self.Encode()
	k := Hash(d)
	Put(headprfx, k, Join([][]byte{
		d,
		boolbytes(signs),
		boolbytes(block),
		PutUint64(uint64(checktime)),
	}))
	return k
}

func GetHeader(key []byte) (*Header, bool, bool, int64, error) { //
	blob := Get(headprfx, key)
	if blob != nil {
		l := Listblob(Split(blob))
		return DecodeHeader(l.Get(0)), l.GetBool(1), l.GetBool(2), int64(l.Getuint64(3)), nil
	}

	blob = Get("tmp-", key)
	if blob != nil {
		header := DecodeHeader(blob)
		if bytes.Equal(header.Hash(), key) {
			s := header.Check()
			header.Cache(s, false, 0)
			return header, s, false, 0, nil
		}
	}

	GetStackByRoot(key)
	blob = Get("tmp-", key)
	if blob != nil {
		header := DecodeHeader(blob)
		if bytes.Equal(header.Hash(), key) {
			s := header.Check()
			header.Cache(s, false, 0)
			return header, s, false, 0, nil
		}
	}
	return nil, false, false, 0, errdata
}

func GetStackByRoot(key []byte) {
	HandleStack(Peers.Request(Join([][]byte{[]byte("get stack"), key}),
		func(blob []byte) bool {
			return bytes.Equal(DecodeHeader(Listblob(Split(blob)).Get(0)).Hash(), key)
		}), key)
}

func (self StateMap) Cache() []byte {
	d := self.Encode()
	k := Hash(d)
	Put(stateprfx, k, d)
	return k
}
func (self BState) Cache() []byte {
	k := Hash(self)
	Put(stateprfx, k, self)
	return k
}

func GetState(key []byte) (BState, error) {
	b := Get(stateprfx, key)
	if b != nil {
		return BState(b), nil
	}

	b = Get("tmp-", key)

	if b != nil {
		res := BState(b)
		res.Cache()
		return res, nil
	}

	b = GetFromNet(stateprfx, key, func(bts []byte) bool {
		return bytes.Equal(Hash(bts), key)
	})
	if b != nil {
		res := BState(b)
		res.Cache()
		return res, nil
	}
	return nil, errdata
}

func (self TxsList) Cache() []byte {
	b := self.Encode()
	k := Hash(b)
	Put(txsprfx, k, b)
	return k
}

func GetTxsList(key []byte) (TxsList, error) {
	b := Get(txsprfx, key)
	if b != nil {
		return DecodeTxsList(b), nil
	}

	b = Get("tmp-", key)

	if b != nil {
		res := DecodeTxsList(b)
		res.Cache()
		return res, nil
	}

	b = GetFromNet(txsprfx, key, func(bts []byte) bool {
		return bytes.Equal(Hash(DecodeTxsList(bts).Encode()), key)
	})

	if b != nil {
		res := DecodeTxsList(b)
		res.Cache()
		return res, nil
	}

	return nil, errdata
}
