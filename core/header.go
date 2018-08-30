package core

import (
	"bytes"
	//"encoding/hex"
	"github.com/triforcecash/triforcecash/core/sign"
	"math/big"
	"time"
)

const (
	signsneed = 2
)

var f64 = new(big.Int).SetBytes([]byte{255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255})

func NewHeader() *Header {
	return &Header{
		Pubs:   make([][]byte, signsneed),
		Signs:  make([][]byte, signsneed),
		Nonces: make([][]byte, signsneed),
	}
}

func coef(b []byte) *big.Int {
	k := new(big.Int).SetBytes(Hash(b))
	k.Div(f64, k)
	return k
}

func (self *Header) Rate() *big.Int {
	if len(self.Pubs) != signsneed || len(self.Nonces) != signsneed {
		return new(big.Int).SetInt64(0)
	}
	var buf bytes.Buffer
	bx := uint64bytes(self.Id)
	buf.Write(bx)
	buf.Write(self.Pubs[0])
	buf.Write(self.Pubs[1])
	x := coef(buf.Bytes())
	var buf1 bytes.Buffer
	buf1.Write(self.Pubs[0])
	buf1.Write(self.Nonces[0])
	var buf2 bytes.Buffer
	buf2.Write(self.Pubs[1])
	buf2.Write(self.Nonces[1])
	k0 := coef(buf1.Bytes())
	k1 := coef(buf2.Bytes())
	x.Mul(x, k0)
	x.Mul(x, k1)
	x.Sqrt(x)
	//	x.Sqrt(x)
	return x
}

func (self *Header) data() []byte {
	var b bytes.Buffer
	b.Write(self.Prev)
	b.Write(self.State)
	b.Write(self.Txs)
	b.Write(uint64bytes(self.Id))
	b.Write(uint64bytes(self.Fee))
	for _, e := range self.Pubs {
		b.Write(e)
	}

	for _, e := range self.Nonces {
		b.Write(e)
	}

	return b.Bytes()
}

func (self *Header) Sign(priv []byte) bool {

	Signmux.Lock()
	defer Signmux.Unlock()
	if !self.SignTokenIsUsed() {

		sig, pub := sign.GenSign(self.data(), priv)

		for i, p := range self.Pubs {
			if bytes.Equal(p, pub) {
				self.Signs[i] = sig
			}
		}
		self.TakeSignToken()
		return true
	}
	return false
}

func (self *Header) Check() bool {
	d := self.data()
	if len(self.Encode()) > headermaxlen || len(self.Pubs) != signsneed || len(self.Signs) != signsneed || len(self.Nonces) != signsneed {
		return false
	}

	for i, s := range self.Signs {
		if !sign.VerSign(d, s, self.Pubs[i]) {
			return false
		}
	}
	return true
}

func (self *Header) Hash() []byte {
	return Hash(self.Encode())
}

func (self *Header) GetPrev() *Header {
	h, s, _, _, err := GetHeader(self.Prev)
	if err != nil {
		return nil
	}
	s = h.Check()
	if !s {
		return nil
	}
	if IsPrev(h, self) {
		return h
	}
	return nil
}

func (self *Header) FullCheck() bool {
	_, s, b, t, _ := GetHeader(self.Hash())
	if !b {
		if t+checktimeout < time.Now().Unix() {
			txs, err := GetTxsList(self.Txs)
			if err != nil {
				self.Cache(s, false, time.Now().Unix())
				return false
			}

			blk := &Block{
				Head: self,
				Txs:  txs,
			}
			res := blk.Check()
			if res {
				s = true
			}
			self.Cache(s, res, time.Now().Unix())

			return res
		}
	}
	return b
}

func (self *Header) Next() *Header {
	return &Header{
		Prev:   self.Hash(),
		Id:     self.Id + 1,
		Pubs:   make([][]byte, signsneed),
		Signs:  make([][]byte, signsneed),
		Nonces: make([][]byte, signsneed),
	}
}

func IsPrev(a, b *Header) bool {
	return a.Id+1 == b.Id && bytes.Equal(a.Hash(), b.Prev)
}

func (self *Header) Back(n int) *Header {
	if n <= 0 {
		return self
	}
	h := self
	for n > 0 {
		h = h.GetPrev()
		if h == nil {
			return nil
		}
		n--
	}
	return h
}

func (self *Header) Fork() *Header {
	tmp0 := make([]byte, len(self.Prev))
	tmp1 := make([]byte, len(self.State))
	tmp2 := make([]byte, len(self.Txs))
	copy(tmp0, self.Prev)
	copy(tmp1, self.State)
	copy(tmp2, self.Txs)
	return &Header{
		Prev:   tmp0,
		State:  tmp1,
		Txs:    tmp2,
		Id:     self.Id,
		Fee:    self.Fee,
		Pubs:   make([][]byte, 2),
		Signs:  make([][]byte, 2),
		Nonces: make([][]byte, 2),
	}
}
func (self *Header) SignKey() []byte {
	if len(self.Pubs) != signsneed {
		return nil
	}
	b := uint64bytes(self.Id)
	b = append(b, self.Pubs[0]...)
	b = append(b, self.Pubs[1]...)
	return Hash(b)
}

func (self *Header) CheckFraud() {
	tb := Get(signtokenprfx, self.SignKey())
	if tb != nil && !bytes.Equal(tb, []byte("empty")) {
		if !bytes.Equal(tb, self.Hash()) {
			b, _, _, _, _ := GetHeader(tb)
			if b != nil {
				CheckFraud(self, b)
			}
		}
	} else {
		Put(signtokenprfx, self.SignKey(), self.Hash())
	}
}

func CheckFraud(a, b *Header) {
	if bytes.Equal(a.SignKey(), b.SignKey()) && !bytes.Equal(a.Hash(), b.Hash()) && a.Check() && b.Check() {
		proof := Join([][]byte{
			a.Encode(),
			b.Encode(),
		})
		Put(banpubprfx, a.Pubs[0], proof)
		Put(banpubprfx, a.Pubs[1], proof)
	}
}

func (self *Header) PublicKeysAreBanned() bool {
	if len(self.Pubs) != signsneed {
		return true
	}
	if Get(banpubprfx, self.Pubs[0]) != nil {
		return true
	}
	if Get(banpubprfx, self.Pubs[1]) != nil {
		return true
	}
	return false
}

func (self *Header) SignTokenIsUsed() bool {
	tb := Get(signtokenprfx, self.SignKey())
	return tb != nil
}

func (self *Header) TakeSignToken() {
	Put(signtokenprfx, self.SignKey(), []byte("empty"))
}
