package core

import (
	"math/big"
	"sync"
	"time"
)

type ChainsPool struct {
	Chains     map[string]*Chain
	Mux        sync.Mutex `json:"-"`
	Difficulty *big.Int
	Main       *Chain
}

func (self *ChainsPool) IncreaseDifficulty() {
	self.Difficulty.Mul(self.Difficulty, big.NewInt(101))
	self.Difficulty.Div(self.Difficulty, big.NewInt(100))
}

func (self *ChainsPool) DecreaseDifficulty() {
	self.Difficulty.Mul(self.Difficulty, big.NewInt(99))
	self.Difficulty.Div(self.Difficulty, big.NewInt(100))
	if self.Difficulty.Cmp(big.NewInt(1000)) == -1 {
		self.Difficulty.SetInt64(1000)
	}
}

func (self *ChainsPool) Add(h *Header) {

	if !h.Check() {
		return
	}

	if h.Rate().Cmp(Candidates.Difficulty) == -1 {
		return
	}
	h.CheckFraud()
	if h.PublicKeysAreBanned() {
		return
	}
	if CurrentId() < h.Id {
		return
	}
	key := string(h.Hash())
	self.Mux.Lock()
	defer self.Mux.Unlock()
	if _, ok := self.Chains[key]; !ok {
		c := &Chain{
			Higher: h,
			Avr:    big.NewInt(0),
			Valid:  true,
		}
		self.Chains[key] = c
		go c.Rate()
		h.Cache(h.Check(), false, 0)
	}
}

func (self *ChainsPool) DeleteLowRate() {
	self.Mux.Lock()
	defer self.Mux.Unlock()
	for k, v := range self.Chains {
		if v.Rate().Cmp(self.Difficulty) == -1 {
			delete(self.Chains, k)
		}
	}
}

func (self *ChainsPool) Update() {
	Candidates.Sign()
	Candidates.Map(func(header *Header) {
		self.Add(header)
	})
	if len(self.Chains) > 100 {
		self.IncreaseDifficulty()
	} else {
		self.DecreaseDifficulty()
	}

	self.Mux.Lock()
	defer self.Mux.Unlock()

	for key, chain := range self.Chains {
		rate := chain.Rate()
		if rate.Cmp(self.Difficulty) == -1 {
			delete(self.Chains, key)
		}
		if Main == nil || !Main.Valid || rate.Cmp(Main.Rate()) == 1 {
			if Main != chain {
				Main = chain
				go Main.StartFullCheck()
			}
		}
	}
}

func (self *Chain) Rate() *big.Int {
	currentid := CurrentId()
	if currentid < self.Higher.Id {
		return big.NewInt(0)
	}

	if self.L == currentid {
		return self.Avr
	}

	self.L = currentid

	penalty := currentid - self.Higher.Id
	h := self.Higher
	num := big.NewInt(int64(penalty))
	sum := big.NewInt(0)
	depth := Checkdepth - int(penalty)

	if depth <= 0 {
		return big.NewInt(0)
	}

	for d := 0; d < depth; d++ {

		num.Add(num, one)
		sum.Add(sum, h.Rate())
		if h.Id <= 0 {
			self.Avr.Div(sum, num)
			return self.Avr
		}

		h = h.GetPrev()

		if h == nil {
			self.Valid = false
			self.Avr.Div(sum, num)
			return self.Avr
		}

	}
	self.Avr.Div(sum, num)
	return self.Avr
}

func (self *Chain) StartFullCheck() {
	h := self.Higher
	for d := 0; d < Checkdepth; d++ {
		if Main != self {
			return
		}

		if !h.FullCheck() {
			self.Valid = false
			return
		}
		if h.Id <= 0 {
			return
		}

		h = h.GetPrev()

		if h == nil {
			self.Valid = false
			return
		}

	}
}

func CurrentId() uint64 {
	return uint64((time.Now().Unix() - StartTime) / BlockTime)
}

func CreateNewBlock() {
	curid:=CurrentId()
	if Mineblocks {
		if Main != nil {
			if Main.Higher.Id < curid {
				newblock := &Block{
					Head: Main.Higher.Next(),
					Txs:  Txs.GetTxs(),
				}
				err := newblock.Create()
				if err != nil {
					return
				}
				newblock.Head.Mine()
			}
		} else {
			newblock := NewBlock()
			err := newblock.Create()
			if err != nil {
				return
			}
			newblock.Head.Mine()
		}
	}
}

func Updater() {
	go func() {
		Chains.Update()
		for {
			CreateNewBlock()
			Chains.Update()
			time.Sleep(250 * time.Millisecond)
		}
	}()
}
