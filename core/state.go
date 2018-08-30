package core

import (
	"bytes"
	"fmt"
	"sort"
)

func (s *State) Encode() []byte {
	b := make([]byte, statelen)
	copy(b[0:32], s.Addr)
	copy(b[32:40], uint64bytes(s.Balance))
	copy(b[40:48], uint64bytes(s.Nonce))
	return b
}

func DecodeState(b []byte) *State {
	tmpblob := make([]byte, statelen)
	copy(tmpblob, b)

	return &State{
		Addr:    tmpblob[0:32],
		Balance: bytesuint64(tmpblob[32:40]),
		Nonce:   bytesuint64(tmpblob[40:48]),
	}
}

func (stm StateMap) Sum() uint64 {
	var sum uint64
	for _, v := range stm {
		sum += v.Balance
	}
	return sum
}

func (stm StateMap) Select(addrs []string) StateMap {
	ns := make(StateMap)
	var ok bool
	for _, addr := range addrs {
		ns[addr], ok = stm[addr]
		if !ok {
			ns[addr] = &State{Addr: []byte(addr)}
		}
	}
	return ns
}

func (stm StateMap) Delete(addrs []string) {
	for _, addr := range addrs {
		delete(stm, addr)
	}
}

func (stm StateMap) Add(stm0 StateMap) {
	for k, s0 := range stm0 {
		v, ok := stm[k]
		if v == nil && !ok {
			stm[k] = s0
		}
	}
}

func (stm StateMap) DeleteIf(f func(value *State) bool) {
	for key, val := range stm {
		if f(val) {
			delete(stm, key)
		}
	}
}

func (stm StateMap) String() string {
	var s string
	for _, v := range stm {
		s += fmt.Sprintf("Addr: %64x Balance: %10d  Nonce: %3d\n", v.Addr, v.Balance, v.Nonce)
	}
	return s

}

type States []*State

func (slc States) Sort() {

	sort.Slice(slc, func(i, j int) bool { return bytes.Compare(slc[i].Addr, slc[j].Addr) < 0 })
}

type BState []byte

func (bst BState) Len() int {
	return len(bst) / statelen
}

func (bst BState) get(idx int) []byte {
	return bst[idx*statelen : (idx+1)*statelen]
}

func (bst BState) Addrs() []string {
	l := bst.Len()
	tmpaddrs := make([]string, l)
	for i := 0; i < l; i++ {
		tmpaddrs[i] = string(bst.get(i)[0:32])
	}
	return tmpaddrs
}
