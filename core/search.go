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
				s.LastBlockId = h.Id
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
