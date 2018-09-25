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
var zero = big.NewInt(0)

func NewChain(h *Header) {
	if Difficulty.Cmp(h.Rate()) == 1 {
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
			Avr:    big.NewInt(0),
			Valid:  true,
		}
		Chains[key] = c
		go c.Rate(CurrentId())
	}
}

func (self *Chain) Rate(currentid uint64) *big.Int {

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
			return big.NewInt(0)
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

		if chain.Valid && (mx == nil || chain.Rate(curid).Cmp(mx.Rate(curid)) == 1) {
			mx = chain
		}

		if mn == nil || chain.Rate(curid).Cmp(mn.Rate(curid)) == -1 {
			mn = chain
			mnk = key
		}

		chainsmux.Lock()
	}
	chainsmux.Unlock()

	if len(Chains) > 200 {
		delete(Chains, mnk)
		IncreaseDifficulty()
	}

	if len(Chains) < 100 {
		DecreaseDifficulty()
	}

	if Main != nil && Main.Higher.Id < curid {
		DecreaseDifficulty()
	}

	if Main != mx && mx != nil {
		Main = mx
		go Main.StartFullCheck()
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

func DecreaseDifficulty() {
	Difficulty.Mul(Difficulty, b99)
	Difficulty.Div(Difficulty, b100)
	if Difficulty.Cmp(MinDifficulty) == -1 {
		Difficulty.Set(MinDifficulty)
	}
}

func IncreaseDifficulty() {
	Difficulty.Mul(Difficulty, b101)
	Difficulty.Div(Difficulty, b100)
}
