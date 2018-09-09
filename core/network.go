package core

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

)

func Serve() {
	if FullNode {
		http.HandleFunc("/api/txshistory", GetTxsHistoryServ)
	}
	http.HandleFunc("/api/db", DBServer)
	http.HandleFunc("/api/mine", MineServ)
	http.HandleFunc("/api/pushtx", PostTx)
	http.HandleFunc("/api/main", MainChain)
	http.HandleFunc("/api/mainjson", MainChainJson)
	http.HandleFunc("/api/statejson", StateJson)
	http.HandleFunc("/api/pushhost", PostHost)
	http.HandleFunc("/api/hosts", HostsServ)
	http.HandleFunc("/api/send", SendServ)
	http.HandleFunc("/api/genaccount", GenAccountServ)
	http.HandleFunc("/api/updatenonce", UpdateNonce)
	http.HandleFunc("/api/updatenoncehex", UpdateNonceHex)
	http.HandleFunc("/", func(res http.ResponseWriter, req *http.Request) {})

	http.ListenAndServe(Port, nil)
}

func HostsServ(res http.ResponseWriter, req *http.Request) {
	res.Header().Set("Content-Type", "application/json")
	hostsmux.Lock()
	b, _ := json.Marshal(Hosts)
	hostsmux.Unlock()

	res.Write(b)
}

func Network() {

	if !ClientOnly {
		go Serve()
	}

	go func() {
		for {

			var buf bytes.Buffer
			if !ClientOnly {
				myhost := &Host{
					Addr:  PublicIp,
					Port:  Port,
					Prot:  protocol,
					Pub:   Pub,
					Nonce: Nonce,
				}
				myhost.Sign()

				b, _ := json.Marshal(myhost)

				buf.Write(b)

			}

			var NewHosts = map[string]*Host{}

			MapHosts(func(url string, host *Host) {
				if host.Karma < (-20) {
					IgnoreAddress(host.Addr)
					return
				}
				if host.Karma > (20) {
					host.Karma = 20
				}

				res, err := http.Post(url+"/api/pushhost", "application/json", &buf)
				if err != nil {
					host.Karma -= 1
					return
				}

				resblob, _ := ioutil.ReadAll(res.Body)
				res.Body.Close()

				var newhosts map[string]*Host
				json.Unmarshal(resblob, &newhosts)
				for k, v := range newhosts {
					NewHosts[k] = v
				}
			})

			for _, host := range NewHosts {
				AddHost(host)
			}
			
			CalculateChanceToCreateBlock()

			time.Sleep(90 * time.Second)
		}
	}()

	go func() {
		for {
			time.Sleep(1800 * time.Second)
			hostsignoremux.Lock()
			HostsIgnore = map[string]int{"127.0.0.1": 0, "0.0.0.0": 0, "255.255.255.255": 0}
			hostsignoremux.Unlock()
		}
	}()
}

func PostHost(res http.ResponseWriter, req *http.Request) {
	hostsmux.Lock()
	bh, _ := json.Marshal(Hosts)
	hostsmux.Unlock()
	res.Write(bh)

	b, _ := ioutil.ReadAll(req.Body)
	req.Body.Close()

	host := &Host{}
	json.Unmarshal(b, host)
	host.Addr = strings.Split(req.RemoteAddr, ":")[0]
	UpdateHost(host)
}

func DBServer(res http.ResponseWriter, req *http.Request) {
	hexkey := req.URL.Query().Get("key")
	key, _ := hex.DecodeString(hexkey)
	data := Get(key, []byte{})
	res.Write(data)
}

func StateJson(res http.ResponseWriter, req *http.Request) {
	hexkey := req.URL.Query().Get("key")
	key, _ := hex.DecodeString(hexkey)
	b, _ := json.Marshal(GetBalance(string(key)))
	res.Header().Set("Content-Type", "application/json")
	res.Write(b)
}

func MainChainJson(res http.ResponseWriter, req *http.Request) {
	if Main != nil {
		res.Header().Set("Content-Type", "application/json")
		b, _ := json.Marshal(Main)
		res.Write(b)
	}
}

func MainChain(res http.ResponseWriter, req *http.Request) {
	if Main != nil {
		res.Header().Set("Content-Type", "application/octet-stream")
		res.Write(Main.Higher.Encode())
	}
}

func NetGet(prfx, key0 []byte, hand func(b, k []byte) bool) []byte {
	var key1 []byte
	key1 = append(key1, prfx...)
	key1 = append(key1, key0...)
	hexkey := hex.EncodeToString(key1)
	hostsmux.Lock()
	for _, host := range Hosts {
		hostsmux.Unlock()
		resp, err := http.Get(host.Prot + host.Addr + host.Port + dbapi + hexkey)
		if resp != nil && err == nil {
			respblob, _ := ioutil.ReadAll(resp.Body)
			resp.Body.Close()
			if hand(respblob, key0) {
				host.Karma += 1
				return respblob
			}
		} else {
			host.Karma -= 1
			log.Println(err)
		}
		hostsmux.Lock()
	}
	hostsmux.Unlock()
	return nil
}

func PostTx(res http.ResponseWriter, req *http.Request) {
	b, err := ioutil.ReadAll(req.Body)
	req.Body.Close()
	if err != nil {
		return
	}
	tx := DecodeTx(b)
	if tx.Check() {
		poolmux.Lock()
		TxsPool = append(TxsPool, tx)
		poolmux.Unlock()
	}
}

func PushTx(tx *Tx) {
	var buf bytes.Buffer
	buf.Write(tx.Encode())

	if tx.Check() {
		poolmux.Lock()
		TxsPool = append(TxsPool, tx)
		poolmux.Unlock()
	}

	MapHosts(func(url string, h *Host) {
		http.Post(url+apipushtx, "application/octet-stream", &buf)
	})

}

func MapHosts(f func(url string, h *Host)) {
	hostsmux.Lock()
	for _, host := range Hosts {
		hostsmux.Unlock()
		f(host.Prot+host.Addr+host.Port, host)
		hostsmux.Lock()
	}
	hostsmux.Unlock()
}

func SendServ(res http.ResponseWriter, req *http.Request) {
	b, err := ioutil.ReadAll(req.Body)
	req.Body.Close()
	if err != nil {
		return
	}
	txreq := &struct {
		Seed   []byte
		Addr   []byte
		Amount uint64
		Fee    uint64
	}{}

	err = json.Unmarshal(b, txreq)
	if err != nil {
		return
	}

	tx := NewTx([][]byte{GenPub(txreq.Seed)}, 1)
	tx.AddOut(string(txreq.Addr), txreq.Amount)
	tx.Fee = txreq.Fee
	s := GetBalance(tx.Sender())
	if s.Balance >= tx.Amount() {
		tx.Nonce = s.Nonce
		tx.Sign(txreq.Seed)
		PushTx(tx)

	}
}

func GenAccountServ(res http.ResponseWriter, req *http.Request) {
	b, err := ioutil.ReadAll(req.Body)
	req.Body.Close()
	if err != nil {
		return
	}
	request := &struct {
		Seed []byte
	}{}

	err = json.Unmarshal(b, request)
	if err != nil {
		return
	}
	response := &struct {
		Pub  []byte
		Addr []byte
	}{
		Pub:  GenPub(request.Seed),
		Addr: []byte(GenAccount(request.Seed)),
	}

	b, err = json.Marshal(response)
	if err != nil {
		return
	}
	res.Write(b)

}

var UpdateNonceMux sync.Mutex

func UpdateNonce(res http.ResponseWriter, req *http.Request) {
	nonce, err := ioutil.ReadAll(req.Body)
	req.Body.Close()
	if err != nil {
		return
	}

	UpdateNonceMux.Lock()
	defer UpdateNonceMux.Unlock()

	if bytes.Compare(Hash(append(Pub, Nonce...)), Hash(append(Pub, nonce...))) == 1 {
		Nonce = nonce
	}

}
func UpdateNonceHex(res http.ResponseWriter, req *http.Request) {
	hexnonce := req.URL.Query().Get("nonce")

	nonce, err := hex.DecodeString(hexnonce)

	if err != nil {
		return
	}

	UpdateNonceMux.Lock()
	defer UpdateNonceMux.Unlock()

	if bytes.Compare(Hash(append(Pub, Nonce...)), Hash(append(Pub, nonce...))) == 1 {
		Nonce = nonce
		res.Write([]byte("true"))
	} else {
		res.Write([]byte("false"))
	}

}

func GetTxsHistoryServ(res http.ResponseWriter, req *http.Request) {

	addrhex := req.URL.Query().Get("key")
	addrb, err := hex.DecodeString(addrhex)
	if err != nil {
		return
	}
	addr := string(addrb)

	blob, err := json.Marshal(GetTxsHistory(addr))
	if err != nil {
		return
	}
	res.Header().Set("Content-Type", "application/json")

	res.Write(blob)
}

