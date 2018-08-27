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
	b := Get(headprfx, key)
	if b != nil {
		args := Listblob(Blob(b).Split())
		return DecodeHeader(args.Get(0)), args.GetBool(1), args.GetBool(2), int64(args.Getuint64(3)), nil
	} else {
		b0 := NetGet(headprfx, key, func(bts, k []byte) bool {
			args := Listblob(Blob(bts).Split())
			return bytes.Equal(DecodeHeader(args.Get(0)).Hash(), k)
		})
		if b0 != nil {
			args := Listblob(Blob(b0).Split())
			h := DecodeHeader(args.Get(0))
			s := h.Check()
			h.CheckFraud()
			h.Cache(s, false, 0)
			return h, s, false, 0, nil
		}
	}
	return nil, false, false, 0, errdata
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
	} else {
		b0 := NetGet(stateprfx, key, func(bts, k []byte) bool {
			return bytes.Equal(Hash(bts), k)
		})
		if b0 != nil {
			res := BState(b0)
			res.Cache()
			return res, nil
		}
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
	} else {
		b0 := NetGet(txsprfx, key, func(bts, k []byte) bool {
			return bytes.Equal(Hash(DecodeTxsList(bts).Encode()), k)
		})
		if b0 != nil {
			res := DecodeTxsList(b0)
			res.Cache()
			return res, nil
		}
	}
	return nil, errdata
}
