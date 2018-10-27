package core

import (
	"log"
	"time"
	"net"
)

func ListenTCP() {
	ln, err := net.Listen("tcp", Port)
	ErrorHandler(err)
	defer ln.Close()
		go func(){for{
			log.Println(s,f)
			time.Sleep(10*time.Second)
			}}()
	for {
		conn, err := ln.Accept()
		ErrorHandler(err)
		HandleConn(conn)
	}
}

func HandleConn(conn net.Conn) {
	if conn == nil {
		return
	}
	if Peers.ConnIsExist(IP(conn.RemoteAddr().String())) {
		conn.Close()
	} else {
		go HandlePeer(NewPeer(conn))
	}
}

func GetFromNet(prfx string, k []byte, hand func(b []byte) bool) []byte {
	key := append([]byte(prfx), k...)
	b := Peers.Request(Join([][]byte{
		[]byte("get"),
		key,
	}),
		func(blob []byte) bool {
			log.Println(hand(blob))
			return hand(blob)
		},
	)
	log.Printf("%q", b)
	return b
}
