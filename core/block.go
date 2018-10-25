package core

import (
	//"log"
	"bytes"
	"errors"
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
