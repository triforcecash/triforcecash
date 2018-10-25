package core

import "bytes"

func (self *Header) Search(res, req StateMap, d int) error {
	h := self
	for c := 0; c < d; c++ {

		if len(req) == 0 {
			break
		}

		if h.Id <= 0 {
			break
		}

		h = h.GetPrev()

		if h == nil {
			return errdata
		}

		st, err := GetState(h.State)

		if st == nil {
			return err
		}

		l := st.Len()

		if l <= 0 {
			continue
		}

		if len(req) == 0 {
			break
		}

		for addr, _ := range req {
			s := st.Search(addr, l)
			if s != nil {
				s.Confirm = self.Id - h.Id
				res[addr] = s
				delete(req, addr)
			}
		}
	}
	return nil
}

func (bst BState) Search(addr string, j int) *State {
	baddr := []byte(addr)
	i := 0
	for i < j {
		h := int(uint(i+j) >> 1)
		switch bytes.Compare(baddr, bst.get(h)[0:32]) {
		case 0:
			return DecodeState(bst.get(h))
		case 1:
			i = h + 1
		case -1:
			j = h
		}

	}
	return nil
}

func (self *Header) IterHeaders(f func(head *Header)) error {
	h := self
	for c := 0; c < period; c++ {

		if h.Id <= 0 {
			break
		}

		h = h.GetPrev()

		if h == nil {
			return errdata
		}

		f(h)
	}
	return nil
}

func (self *Tx) AddrMatch(addr string) bool {
	for _, addr0 := range self.Addrs() {
		if addr0 == addr {
			return true
		}
	}
	return false
}

func (self TxsList) SearchByAddr(addr string) TxsList {
	res := TxsList{}
	for _, tx := range self {
		if tx.AddrMatch(addr) {
			res = append(res, tx)
		}
	}
	return res
}

type SearchTxsResultItem struct {
	State   *State
	Header  *Header
	TxsList TxsList
}

func (self *Header) SearchTxs(addr string) []SearchTxsResultItem {
	res := []SearchTxsResultItem{}
	self.IterHeaders(
		func(head *Header) {
			state, err := GetState(head.State)

			if state == nil || err != nil {
				return
			}

			s := state.Search(addr, state.Len())

			if s == nil {
				return
			}
			s.Confirm = self.Id - head.Id
			txslist, _ := GetTxsList(head.Txs)
			if txslist == nil {
				return
			}

			PrevHeaderReward := head.GetPrev()

			if PrevHeaderReward != nil && (!(Addr(PrevHeaderReward.Pubs[0]) == addr || Addr(PrevHeaderReward.Pubs[1]) == addr)) {
				PrevHeaderReward = nil
			}

			res = append(res, SearchTxsResultItem{
				State:   s,
				Header:  PrevHeaderReward,
				TxsList: txslist.SearchByAddr(addr),
			})
		})
	return res
}
