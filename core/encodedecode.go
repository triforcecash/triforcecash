package core

import "bytes"

func (t *Tx) Encode() []byte {
	return Join([][]byte{
		Join(t.Pubs),
		Join(t.Signs),
		[]byte{t.Needvotes},
		Join(t.Outs),
		PutUint64(t.Fee),
		PutUint64(t.Nonce),
		PutUint64(t.TimeLock),
		t.Hash,
		t.Proof,
	})
}

func DecodeTx(bts []byte) *Tx {
	b := Listblob(Split(bts))
	return &Tx{
		Pubs:      Split(b.Get(0)),
		Signs:     Split(b.Get(1)),
		Needvotes: b.Getuint8(2),
		Outs:      Split(b.Get(3)),
		Fee:       b.Getuint64(4),
		Nonce:     b.Getuint64(5),
		TimeLock:  b.Getuint64(6),
		Hash:      b.Get(7),
		Proof:     b.Get(8),
	}
}

func (self TxsList) Encode() []byte {
	encodedtxs := make([][]byte, len(self))
	for i, t := range self {
		encodedtxs[i] = t.Encode()
	}
	return Join(encodedtxs)
}

func DecodeTxsList(b []byte) TxsList {
	encodedtxs := Split(b)
	txs := make(TxsList, len(encodedtxs))
	for i, bt := range encodedtxs {
		txs[i] = DecodeTx(bt)
	}
	return txs
}

func (self *Header) Encode() []byte {
	return Join([][]byte{
		self.Prev,
		self.State,
		self.Txs,
		PutUint64(self.Id),
		PutUint64(self.Fee),
		Join(self.Pubs),
		Join(self.Signs),
		Join(self.Nonces),
	})
}

func DecodeHeader(b []byte) *Header {
	args := Listblob(Split(b))
	return &Header{
		Prev:   args.Get(0),
		State:  args.Get(1),
		Txs:    args.Get(2),
		Id:     args.Getuint64(3),
		Fee:    args.Getuint64(4),
		Pubs:   Split(args.Get(5)),
		Signs:  Split(args.Get(6)),
		Nonces: Split(args.Get(7)),
	}
}

func (self *Block) Encode() []byte {
	return Join([][]byte{
		self.Head.Encode(),
		self.Txs.Encode(),
	})
}

func DecodeBlock(b []byte) *Block {
	args := Listblob(Split(b))
	return &Block{
		Head: DecodeHeader(args.Get(0)),
		Txs:  DecodeTxsList(args.Get(1)),
	}
}

func (stm StateMap) Encode() BState {
	l := len(stm)

	tmpslice := make(States, l)
	i := 0
	for _, stt := range stm {
		tmpslice[i] = stt
		i++
	}
	tmpslice.Sort()
	return BState(tmpslice.Encode())
}

func DecodeStateMap(self BState) StateMap {
	l := self.Len()
	sm := make(StateMap)
	for i := 0; i < l; i++ {
		s := DecodeState(self.get(i))
		sm[string(s.Addr)] = s
	}
	return sm
}

func (slc States) Encode() []byte {
	var blob bytes.Buffer
	for _, stt := range slc {
		blob.Write(stt.Encode())
	}
	return blob.Bytes()
}
