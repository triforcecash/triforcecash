package core

import (
	"bytes"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"github.com/triforcecash/triforcecash/core/sign"
	"log"
	"os"
	"time"
)

func GetBalance(addr string) *State {
	if Main != nil {
		req := (StateMap{}).Select([]string{addr})
		res := (StateMap{})
		Main.Higher.Search(res, req, period)
		res = res.Select([]string{addr})
		s := res[addr]
		return s
	} else {
		return nil
	}
}

func GetTxsHistory(addr string) []SearchTxsResultItem {
	if Main != nil {
		return Main.Higher.SearchTxs(addr)
	}
	return nil
}

func MineKey() {
	for {

		if !Minecpu || Pub == nil {
			time.Sleep(60 * time.Second)
			continue
		}

		mn := Hash(append(Pub, Nonce...))
		nonce := make([]byte, 32)
		rand.Read(nonce)
		for i := 0; i < 10000000; i++ {
			n := Hash(append(Pub, nonce...))
			if bytes.Compare(mn, n) == 1 {
				Nonce = nonce
				mn = n
			}
			nonce = n

		}
	}
}

func SaveNonce() {
	if Pub == nil {
		return
	}
	f, err := os.Create(fmt.Sprintf("accounts/%x", Pub))
	if err != nil {
		log.Println(err)
		os.Mkdir("accounts/", 0777)
		f, err := os.Create(fmt.Sprintf("accounts/%x", Pub))
		if err != nil {
			return
		}
		f.Write(Nonce)
		f.Close()

	}
	f.Write(Nonce)
	f.Close()

}

func LoadNonce() {
	f, err := os.Open(fmt.Sprintf("accounts/%x", Pub))
	if err != nil {
		log.Println(err)
	} else {
		st, err1 := f.Stat()
		if err1 != nil {
			log.Println(err1)
		} else {
			size := st.Size()
			nonce := make([]byte, size)
			f.Read(nonce)
			f.Close()
			Nonce = nonce
		}
	}
}

func SetSeed(seed []byte) ([]byte, []byte, []byte) {
	SaveNonce()
	Priv = seed
	priv := sign.PrivateKeyFromSeed(seed)
	_, Pub = sign.GenSign([]byte{}, seed)
	addr := Addr(Pub)
	LoadNonce()
	return []byte(addr), priv, Pub

}

func Load() {
	b := Get("", []byte("chains"))
	if b != nil {
		//json.Unmarshal(b, &Chains)
	}
}

func Save() {

	b, _ := json.Marshal(Chains)

	Put("", []byte("chains"), b)

	SaveNonce()
}

func Start() {
	RemoveTmp()
	Load()
	Updater()
	go MineKey()
	go func() {
		for {
			time.Sleep(300 * time.Second)
			Save()
		}
	}()

	go func() {
		for {
			Keys.Sync()
			time.Sleep(10 * time.Second)
		}
	}()

	go func(){
		for{
			Peers.Sync()
			time.Sleep(10*time.Second)
		}
		}()

	go func(){
		for{
			Candidates.Sync()
			time.Sleep(10 * time.Second)
		}
		}()

	go func(){
		for{
			Txs.Sync()
			time.Sleep(10*time.Second)
		}
	}()
	Serve()

}

func Stop() {
	Save()
	DBClose()
}
