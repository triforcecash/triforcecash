package core

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"sync"
)

func Serve() {
	if PortHTTP != "" {
		http.HandleFunc("/api/txshistory", GetTxsHistoryServ)
		http.HandleFunc("/api/pushtx", PostTx)
		http.HandleFunc("/api/main", MainChainJson)
		http.HandleFunc("/api/state", StateJson)
		http.HandleFunc("/api/candidates", CandidatesJson)
		http.HandleFunc("/api/txspool", TxsJson)
		http.HandleFunc("/api/keys", KeysJson)
		http.HandleFunc("/api/peers", PeersJson)
		http.HandleFunc("/api/send", SendServ)
		http.HandleFunc("/api/genaccount", GenAccountServ)
		http.HandleFunc("/api/updatenonce", UpdateNonce)
		http.HandleFunc("/api/updatenoncehex", UpdateNonceHex)
		http.HandleFunc("/", ExplorerServ)
		go http.ListenAndServe(PortHTTP, nil)
	}
	Peers.Connect(Lobby)
	go Peers.ConnectToPeers()
	ListenTCP()
}

func DBServer(res http.ResponseWriter, req *http.Request) {
	hexkey := req.URL.Query().Get("key")
	key, _ := hex.DecodeString(hexkey)
	data := Get("", key)
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

func CandidatesJson(res http.ResponseWriter, req *http.Request) {
	res.Header().Set("Content-Type", "application/json")
	b, _ := json.Marshal(Candidates)
	res.Write(b)
}

func TxsJson(res http.ResponseWriter, req *http.Request) {
	res.Header().Set("Content-Type", "application/json")
	b, _ := json.Marshal(Txs)
	res.Write(b)
}

func PeersJson(res http.ResponseWriter, req *http.Request) {
	res.Header().Set("Content-Type", "application/json")
	p := *Peers
	b, err := json.Marshal(p)
	res.Write(b)

}

func KeysJson(res http.ResponseWriter, req *http.Request) {
	res.Header().Set("Content-Type", "application/json")
	b, _ := json.Marshal(Keys)
	res.Write(b)
}

func MainChain(res http.ResponseWriter, req *http.Request) {
	if Main != nil {
		res.Header().Set("Content-Type", "application/octet-stream")
		res.Write(Main.Higher.Encode())
	}
}

func PostTx(res http.ResponseWriter, req *http.Request) {
	b, err := ioutil.ReadAll(req.Body)
	req.Body.Close()
	if err != nil {
		return
	}
	tx := DecodeTx(b)
	Txs.Add(tx)

}

func PushTx(tx *Tx) {

	encodedtx := tx.Encode()
	Txs.Add(tx)

	var buf bytes.Buffer
	buf.Write(encodedtx)
	http.Post(Lobby+apipushtx, "application/octet-stream", &buf)
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
