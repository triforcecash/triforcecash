package core

import (
	"net"
)

func ListenTCP() {
	ln, err := net.Listen("tcp", Port)
	ErrorHandler(err)
	defer ln.Close()
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
			return hand(blob)
		},
	)
	return b
}
