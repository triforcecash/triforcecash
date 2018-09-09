package main

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"flag"
	"golang.org/x/crypto/sha3"
	"log"
	"net/http"
	"time"
)

var Host string
var Pub []byte
var pushchan = make(chan []byte, 1000)

func main() {
	pubhex := flag.String("publickey", "", "Public key hex")
	host := flag.String("host", "", "Node address ip:port ")
	thr := flag.Int("threads", 1, "Number of threads")
	flag.Parse()
	var err error
	Pub, err = hex.DecodeString(*pubhex)
	if err != nil {
		log.Fatal(err)
		return
	}
	Host = *host
	for i := 0; i < *thr; i++ {
		go RunThread()
	}

	for {
		select {
		case nonce := <-pushchan:
			Push(nonce)
		}
	}
}

func Hash(data []byte) []byte {
	tmp := sha3.Sum256(data)
	return tmp[:]
}

func RunThread() {
	mn := Hash(append(Pub, []byte{}...))
	for {
		nonce := make([]byte, 32)
		rand.Read(nonce)
		n := Hash(append(Pub, nonce...))
		if bytes.Compare(mn, n) == 1 {
			log.Printf("%x\n", nonce)
			pushchan <- nonce
			mn = n
		}
		nonce = n
	}
}

func Push(nonce []byte) {
	var buf bytes.Buffer
	buf.Write(nonce)
	res, err := http.Post("http://"+Host+"/api/updatenonce", "application/octet-stream", &buf)

	for err != nil || res.StatusCode != 200 {
		time.Sleep(15 * time.Second)
		res, err = http.Post("http://"+Host+"/api/updatenonce", "application/octet-stream", &buf)
	}

}
