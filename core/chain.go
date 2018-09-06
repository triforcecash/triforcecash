package core

import (
	"io/ioutil"
	"log"
	"math/big"
	"net/http"
	"time"
)

var b100 = new(big.Int).SetInt64(100)
var b99 = new(big.Int).SetInt64(99)
var b101 = new(big.Int).SetInt64(101)

func NewChain(h *Header) {
	if Difficult.Cmp(h.Rate()) == 1 {
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

	chainsmux.Lock()
	defer chainsmux.Unlock()

	_, ok := Chains[key]

	if !ok {
		c := &Chain{
			Higher: h,
			Active: true,
			Valid:  true,
			Avr:    new(big.Int),
		}
		Chains[key] = c
		go c.Start()
	}
}

func (self *Chain) Start() {
	self.Mux.Lock()
	self.Active = true
	self.Mux.Unlock()
	defer func() {
		self.Mux.Lock()
		self.Active = false
		self.Mux.Unlock()
	}()

	h := self.Higher

	num := new(big.Int).SetInt64(0)
	sum := new(big.Int).SetInt64(0)

	for d := 0; d < Checkdepth; d++ {

		num.Add(num, one)
		sum.Add(sum, h.Rate())

		self.Mux.Lock()
		self.Avr.Div(sum, num)
		self.Mux.Unlock()

		if h.Id <= 0 {
			return
		}

		h = h.GetPrev()

		if h == nil {
			self.Mux.Lock()
			self.Valid = false
			self.Mux.Unlock()
			return
		}

	}
}

func (self *Chain) StartFullCheck() {
	h := self.Higher
	for d := 0; d < Checkdepth; d++ {
		if Main != self {
			return
		}

		if !h.FullCheck() {
			self.Mux.Lock()
			self.Valid = false
			self.Avr.SetInt64(0)
			self.Mux.Unlock()
			return
		}
		if h.Id <= 0 {
			return
		}

		h = h.GetPrev()

		if h == nil {
			self.Mux.Lock()
			self.Valid = false
			self.Avr.SetInt64(0)
			self.Mux.Unlock()
			return
		}

	}
}

func CurrentId() uint64 {
	return uint64((time.Now().Unix() - StartTime) / BlockTime)
}

func CreateNewBlock(curid uint64) {

	if Mineblocks {

		if Main != nil {

			if Main.Higher.Id < curid {

				newblock := &Block{
					Head: Main.Higher.Next(),
					Txs:  GetTxsFromPool(),
				}
				err := newblock.Create()
				if err != nil {
					log.Println(err)
					return
				}
				newblock.Mine()
			}
		} else {
			newblock := NewBlock()
			err := newblock.Create()
			if err != nil {
				log.Println(err)
				return
			}
			newblock.Mine()
		}
	}
}

func Update(curid uint64) {
	var mx *Chain
	var mn *Chain
	//var mxk string
	var mnk string

	MapHosts(func(url string, host *Host) {
		res, _ := http.Get(url + apimainchain)
		if res != nil {
			b, _ := ioutil.ReadAll(res.Body)
			res.Body.Close()
			NewChain(DecodeHeader(b))
		}
	})

	chainsmux.Lock()
	for key, chain := range Chains {
		chainsmux.Unlock()
		chain.Mux.Lock()

		if chain.Valid && (mx == nil || chain.Avr.Cmp(mx.Avr) == 1 && chain.Higher.Id == mx.Higher.Id || chain.Higher.Id > mx.Higher.Id && chain.Higher.Id <= curid) {
			mx = chain
		}

		if mn == nil || chain.Avr.Cmp(mn.Avr) == -1 && chain.Higher.Id == mn.Higher.Id || chain.Higher.Id < mn.Higher.Id && mn.Higher.Id <= curid {
			mn = chain
			mnk = key
		}

		chain.Mux.Unlock()
		chainsmux.Lock()
	}
	chainsmux.Unlock()

	if len(Chains) > 100 {
		delete(Chains, mnk)
		Difficult.Mul(Difficult, b101)
		Difficult.Div(Difficult, b100)
	}

	if len(Chains) < 100 {
		Difficult.Mul(Difficult, b99)
		Difficult.Div(Difficult, b100)
		if Difficult.Cmp(MinDifficult) == -1 {
			Difficult.Set(MinDifficult)
		}
	}

	if Main != nil && Main.Higher.Id < curid {
		Difficult.Mul(Difficult, b99)
		Difficult.Div(Difficult, b100)
		if Difficult.Cmp(MinDifficult) == -1 {
			Difficult.Set(MinDifficult)
		}
	}

	if Main != mx && mx != nil {
		Main = mx
		Main.StartFullCheck()
	}
	if Main != mx && mx == nil {
		Main = mx
	}

}

func Updater() {
	go func() {
		Update(CurrentId())
		for {
			cid := CurrentId()
			CreateNewBlock(cid)
			Update(cid)
			time.Sleep(250 * time.Millisecond)
		}
	}()
}
