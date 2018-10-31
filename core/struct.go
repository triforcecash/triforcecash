package core

import (
	"errors"
	"math/big"
	"sync"
)

const (
	maxsupply    = 1e12
	period       = 100000
	txsmaxlen    = 1 << 21
	txmaxlen     = 1 << 12
	headermaxlen = 1 << 12
	Salt         = "test.triforcecash.com"
	StartTime    = 1541009424
	BlockTime    = 90
	checktimeout = 180
	statelen     = 48

	headprfx      = "head-"
	stateprfx     = "state-"
	txsprfx       = "txs-"
	hostprfx      = "host-"
	signtokenprfx = "signtoken-"
	banpubprfx    = "banpub-"
)

var (
	Main    *Chain
	mainmux sync.Mutex

	Lobby      = "triforcecash.com"
	PortHTTP   = ""
	Mineblocks bool
	Minecpu    bool
	Checkdepth = 1000

	Nonce []byte
	Priv  []byte
	Pub   []byte

	Candidates = &CandidatesPool{
		Candidates: make(map[string]*Header),
		Difficulty: big.NewInt(1000),
	}

	Chains = &ChainsPool{
		Chains:     make(map[string]*Chain),
		Difficulty: big.NewInt(1000),
	}

	Keys = &KeysPool{
		Keys:       make(map[string]*Key),
		Difficulty: big.NewInt(1000),
		Total:      big.NewInt(0),
	}

	Peers = &PeersPool{
		Peers:   make(map[string]*Peer),
		Ignored: map[string]bool{"127.0.0.1": false, "0.0.0.0": false, "255.255.255.255": false},
	}

	Txs = &TxsPool{
		Txs: make(map[string]*Tx),
	}

	errsum    = errors.New("sumbefore > sumafter")
	errdata   = errors.New("Can not find data")
	errfatal  = errors.New("Fatal")
	errtxs    = errors.New("Transaction check fail")
	errheader = errors.New("Header signs is invalid")
	errblock  = errors.New("Block is invalid")

	Signmux sync.Mutex

	one = new(big.Int).SetInt64(1)
)

type Chain struct {
	Higher *Header
	Avr    *big.Int
	L      uint64
	Valid  bool
}

type Block struct {
	Head  *Header
	Txs   TxsList
	State StateMap
}

type Header struct {
	Prev   []byte
	State  []byte
	Txs   []byte
	Id     uint64
	Fee    uint64
	Pubs   [][]byte
	Signs  [][]byte
	Nonces [][]byte
}

type TxsList []*Tx

type StateMap map[string]*State

type Tx struct {
	Pubs      [][]byte
	Signs     [][]byte
	Needvotes uint8
	Outs      [][]byte
	Fee       uint64
	Nonce     uint64
	TimeLock  uint64
	Hash      []byte
	Proof     []byte
}

type State struct {
	Addr    []byte
	Balance uint64
	Nonce   uint64
	Confirm uint64
}

type Host struct {
	Addr  string
	Port  string
	Prot  string
	Pub   []byte
	Nonce []byte
	Karma int64
	Part  float64
	Proof []byte
}

func Reward(id uint64) uint64 {
	const k = 1000000
	return ((maxsupply - 6e10) >> (1 + id/k)) / k
}
