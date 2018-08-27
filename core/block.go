package core

import (
	//"log"
	"bytes"
	"errors"
	"io/ioutil"
	"net/http"
)

func NewBlock() *Block {
	return &Block{
		Head:  NewHeader(),
		Txs:   make(TxsList, 0),
		State: make(StateMap),
	}
}

func (self *Block) Create() error {

	var crtprevblck []string
	var prevfee uint64

	if self.Head.Id > 0 {
		prev := self.Head.GetPrev()
		if prev == nil {
			return errors.New("getprev")
		}
		if len(prev.Pubs) != signsneed {
			return errfatal
		}
		crtprevblck = []string{
			Addr(prev.Pubs[0]),
			Addr(prev.Pubs[1]),
		}
		prevfee = prev.Fee
	}
	addrs := append(self.Txs.Addrs(), crtprevblck...) //addrs from tx

	req := (StateMap{}).Select(addrs)
	res := (StateMap{})

	err := self.Head.Search(res, req, period)
	if err != nil {
		return errors.New("Search")
	}

	self.State = res.Select(addrs)

	//reward
	if len(crtprevblck) == signsneed {
		reward := Reward(self.Head.Id) + prevfee
		self.State[crtprevblck[0]].Balance += reward >> 1
		self.State[crtprevblck[1]].Balance += reward >> 1
	}

	sumbefore := self.State.Sum()
	tmptxs := TxsList{}
	for _, t := range self.Txs {
		if t.Transfer(self.State) {
			tmptxs = append(tmptxs, t)
		}
	}
	self.Txs = tmptxs

	sumafter := self.State.Sum()
	if sumafter > sumbefore {
		return errsum
	}
	fee := sumbefore - sumafter

	addrs = append(self.Txs.Addrs(), crtprevblck...)
	self.State = self.State.Select(addrs)

	if self.Head.Id >= period {
		inhblk := self.Head.Back(period)
		if inhblk == nil {
			return errors.New("getback")
		}

		inhstt, err := GetState(inhblk.State)

		if err != nil {
			return err
		}
		req := DecodeStateMap(inhstt)
		req.Delete(addrs)
		res := StateMap{}
		err0 := self.Head.Search(res, req, period-1)

		if err0 != nil {
			return err0
		}
		sumb := req.Sum()
		req.DeleteIf(func(s *State) bool {
			if s.Balance == 0 {
				return true
			} else {
				s.Balance -= 1
				return false
			}
		})
		suma := req.Sum()
		if suma > sumb {
			return errsum
		}
		fee += sumb - suma
		self.State.Add(req)
	}
	if self.Head.Id == 0 {
		//instamine developer fund 1%
		self.State[string(fundaccount)] = &State{Addr: fundaccount, Balance: 1e10, Nonce: 0}
	}
	self.Head.Fee = fee
	self.Head.Txs = self.Txs.Cache()
	self.Head.State = self.State.Cache()
	return nil
}

func (self *Block) Next() *Block {
	return &Block{
		Head:  self.Head.Next(),
		Txs:   make(TxsList, 0),
		State: make(StateMap),
	}
}

func (self *Block) Check() bool {
	
	if !self.Head.Check() || !self.Txs.Check() {
		return false
	}

	headhash := self.Head.Hash()
	txshash := self.Txs.Hash()
	err := self.Create()

	if err != nil {
		return false
	}
	if bytes.Equal(headhash, self.Head.Hash()) && bytes.Equal(txshash, self.Head.Txs) {
		return true
	} else {
		return false
	}
}
func (self *Block) Fork() *Block {
	return &Block{
		Head: self.Head.Fork(),
		Txs:  self.Txs,
	}

}

func (self *Block) Mine() {
	if Pub == nil || Priv == nil || Nonce == nil {
		return
	}
	b := self.Fork()
	b.Head.Pubs = [][]byte{Pub, Pub}
	b.Head.Nonces = [][]byte{Nonce, Nonce}
	if b.Head.Rate().Cmp(Difficult) == 1 {

		if !b.Head.SignTokenIsUsed() {
			b.Head.Sign(Priv)
			b.Head.Cache(true, true, 0)
			b.Head.CheckFraud()
			NewChain(b.Head)
		}
	}

	MapHosts(func(url string, host *Host) {

		if host.Pub == nil || host.Nonce == nil {
			return
		}

		b := self.Fork()
		b.Head.Pubs = [][]byte{Pub, host.Pub}
		b.Head.Nonces = [][]byte{Nonce, host.Nonce}
		if b.Head.Rate().Cmp(Difficult) == 1 {
			if !b.Head.SignTokenIsUsed() {
				b.Head.Sign(Priv)
				b.Head.TakeSignToken()
				var buf bytes.Buffer
				buf.Write(b.Encode())
				go http.Post(url+mineapi, "application/octet-stream", &buf)
			}
		}

	})
}

func MineServ(res http.ResponseWriter, req *http.Request) {
	defer func() {
		recover()
	}()

	blb, err := ioutil.ReadAll(req.Body)
	req.Body.Close()
	if err != nil {
		return
	}
	blk := DecodeBlock(blb)
	if blk.Head.Rate().Cmp(Difficult) == 1 && Main!=nil && Main.Higher.Id <= blk.Head.Id && blk.Head.Id <= CurrentId(){
		if !blk.Head.SignTokenIsUsed() && !blk.Head.PublicKeysAreBanned() {
			blk.Head.Sign(Priv)
			Put(signtokenprfx, blk.Head.SignKey(), blk.Head.Hash())
			blk.Head.Txs = Hash(blk.Txs.Encode())
			if blk.Head.Check() {
				blk.Txs.Cache()
				blk.Head.Cache(true, false, 0)
				NewChain(blk.Head)
			}
		}
	}
}
