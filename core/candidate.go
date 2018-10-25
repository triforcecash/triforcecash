package core

import (
	"log"
	"math/big"
	"sync"
)

const (
	CandidatesLimit = 100
)

type CandidatesPool struct {
	Candidates map[string]*Header
	Mux        sync.Mutex
	Difficulty *big.Int
}

func (self *CandidatesPool) IncreaseDifficulty() {
	self.Difficulty.Mul(self.Difficulty, big.NewInt(101))
	self.Difficulty.Div(self.Difficulty, big.NewInt(100))
}

func (self *CandidatesPool) DecreaseDifficulty() {
	self.Difficulty.Mul(self.Difficulty, big.NewInt(99))
	self.Difficulty.Div(self.Difficulty, big.NewInt(100))
	if self.Difficulty.Cmp(big.NewInt(1000)) == -1 {
		self.Difficulty.SetInt64(1000)
	}
}

func (self *CandidatesPool) Add(header *Header) {
	if header == nil {
		return
	}

	curid := CurrentId()
	minid := uint64(0)
	if Main != nil {
		minid = Main.Higher.Id
	}

	self.Mux.Lock()
	defer self.Mux.Unlock()

	if header.Id >= minid && header.Id <= curid && header.Rate().Cmp(self.Difficulty) == 1 && len(header.Encode()) <= headermaxlen {
		key := string(header.SignKey())
		if header0, ok := self.Candidates[key]; ok {
			if header0.NumSigns() < header.NumSigns() {
				self.Candidates[key] = header
			}
		} else {
			if header.NumSigns() > 0 {
				self.Candidates[key] = header
			}
		}
	}
}

func (self *CandidatesPool) DeleteLowRate() {

	curid := CurrentId()
	minid := uint64(0)
	if Main != nil {
		minid = Main.Higher.Id
	}

	self.Mux.Lock()
	defer self.Mux.Unlock()

	for key, header := range self.Candidates {
		if !(header.Id >= minid && header.Id <= curid && header.Rate().Cmp(self.Difficulty) == 1 && len(header.Encode()) <= headermaxlen) {
			delete(self.Candidates, key)
		}
	}
}

func (self *CandidatesPool) AddNew(blob []byte) {
	l := Split(blob)
	for _, h := range l {
		self.Add(DecodeHeader(h))
	}
}

func (self *CandidatesPool) Sync() {
	defer func() {
		if r := recover(); r != nil {
			log.Println(r)
		}
	}()
	res := Peers.Action(Join([][]byte{
		[]byte("sync candidates"),
		[]byte{},
	}),
		func(blob []byte) bool {
			return true
		})

	self.AddNew(res)

	if len(self.Candidates) > CandidatesLimit {
		self.IncreaseDifficulty()
	} else {
		self.DecreaseDifficulty()
	}

	self.DeleteLowRate()
}

func (self *CandidatesPool) Encode() []byte {
	var res [][]byte

	self.Mux.Lock()
	defer self.Mux.Unlock()

	for _, v := range self.Candidates {
		res = append(res, v.Encode())
	}

	return Join(res)
}

func (self *CandidatesPool) Map(fun func(head *Header)) {
	self.Mux.Lock()
	for _, header := range self.Candidates {
		self.Mux.Unlock()
		fun(header)
		self.Mux.Lock()
	}
	self.Mux.Unlock()
}

func (self *CandidatesPool) Sign() {
	self.Map(
		func(header *Header) {
			header.Sign()
		})
}
