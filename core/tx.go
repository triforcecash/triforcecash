package core

import (
	"bytes"
	"github.com/triforcecash/triforcecash/core/sign"
	"time"
)

type out []byte

func NewTx(pubs [][]byte, nv uint8) *Tx {
	return &Tx{
		Pubs:      pubs,
		Needvotes: nv,
		Signs:     make([][]byte, len(pubs)),
		Outs:      make([][]byte, 0),
	}

}

func Addr(pub []byte) string {
	return NewTx([][]byte{pub}, 1).Sender()
}

func (t *Tx) Sender() string {
	var buf bytes.Buffer
	for _, pub := range t.Pubs {
		buf.Write(pub)
	}
	buf.Write([]byte{t.Needvotes})
	return string(Hash(buf.Bytes()))
}

func (t *Tx) AddOut(addr string, amount uint64) {
	ot := Join([][]byte{[]byte(addr), PutUint64(amount)})
	t.Outs = append(t.Outs, ot)
}

func (t *Tx) data() []byte {
	var buf bytes.Buffer

	buf.Write([]byte(t.Sender()))

	for _, o := range t.Outs {
		buf.Write(o)
	}

	buf.Write(uint64bytes(t.Fee))
	buf.Write(uint64bytes(t.Nonce))
	buf.Write(uint64bytes(t.TimeLock))
	buf.Write(t.Hash)
	return buf.Bytes()
}

func (t *Tx) Sign(priv []byte) {
	s, pb := sign.GenSign(t.data(), priv)
	for i, pb0 := range t.Pubs {
		if bytes.Equal(pb, pb0) {
			t.Signs[i] = s
		}
	}
}

func (self out) getAddr() string {
	return string(Listblob(Blob(self).Split()).Get(0))
}

func (self out) getAmount() uint64 {
	return Listblob(Blob(self).Split()).Getuint64(1)
}

func (t *Tx) Amount() uint64 {
	sum := t.Fee
	for _, o := range t.Outs {
		sum += out(o).getAmount()
	}
	return sum
}

func (t *Tx) Check() bool {
	if len(t.Encode()) > txmaxlen {
		return false
	}

	if len(t.Pubs) != len(t.Signs) {
		return false
	}

	dt := t.data()
	votes := 0
	for i, _ := range t.Signs {
		if sign.VerSign(dt, t.Signs[i], t.Pubs[i]) {
			votes++
		}
	}

	if int(t.Needvotes) > votes {
		return false
	}

	if bytes.Equal(HashSha256(t.Proof), t.Hash) || int64(t.TimeLock) < time.Now().Unix() {
		return true
	} else {
		return false
	}
	return false
}

func (t *Tx) Transfer(states StateMap) bool {
	amnt := t.Amount()
	sndr := states[t.Sender()]

	if t.Check() && sndr.Balance >= amnt && sndr.Nonce == t.Nonce {
		sndr.Balance -= amnt
		sndr.Nonce++
		for _, o := range t.Outs {
			states[out(o).getAddr()].Balance += out(o).getAmount()
		}
		return true
	} else {
		return false
	}
}

func (self *Tx) Addrs() []string {
	var tmpaddrs []string
	tmpaddrs = append(tmpaddrs, self.Sender())
	for _, o := range self.Outs {
		tmpaddrs = append(tmpaddrs, out(o).getAddr())
	}
	return tmpaddrs
}

func (self TxsList) Addrs() []string {
	var tmpaddrs []string
	for _, trx := range self {
		tmpaddrs = append(tmpaddrs, trx.Sender())
		for _, o := range trx.Outs {
			tmpaddrs = append(tmpaddrs, out(o).getAddr())
		}
	}
	return tmpaddrs

}

func (self TxsList) Check() bool {
	if len(self.Encode()) > txsmaxlen {
		return false
	}
	for _, t := range self {
		if !t.Check() {
			return false
		}
	}
	return true
}

func (self TxsList) Hash() []byte {
	return Hash(self.Encode())
}

func GenAccount(seed []byte) string {
	_, pub := sign.GenSign([]byte{}, seed)
	addr := Addr(pub)
	return addr
}

func GenPub(seed []byte) []byte {
	_, pub := sign.GenSign([]byte{}, seed)
	return pub
}
