package core

import (
	"errors"
	"net"
	"sync"
	"time"
)

const (
	MaxPeersLen = 1000
	Port        = ":6885"
	PortNum     = 6885
	Deadline    = 600
	MaxPeers    = 400
	Delay       = 250
	TimeOut     = 30
	ReadLimit   = 1 << 26
)

type Peer struct {
	RemoteAddr   string
	Conn         net.Conn
	ReadChan     chan []byte `json:"-"`
	TmpChan      chan []byte `json:"-"`
	WriteChan    chan []byte `json:"-"`
	RequestChan  chan []byte `json:"-"`
	ResponseChan chan []byte `json:"-"`
}

type PeersPool struct {
	Peers   map[string]*Peer
	Mux     sync.Mutex
	Num     int
	Ignored map[string]bool
}

var errconn = errors.New("Connection closed")

func (self *PeersPool) NumAdd() {
	self.Mux.Lock()
	defer self.Mux.Unlock()
	self.Num++
}

func (self *PeersPool) Connect(ip string) {

	if _, ok := self.Ignored[ip]; ok {
		return
	}

	if self.ConnIsExist(ip) {
		return
	} else {
		conn, err := net.Dial("tcp", ip+Port)
		if err != nil {
			self.Delete(ip)
		}
		ErrorHandler(err)
		go HandlePeer(NewPeer(conn))
	}
}

func (self *PeersPool) Map(f func(ip string, peer *Peer)) {
	self.Mux.Lock()
	for ip, peer := range Peers.Peers {
		self.Mux.Unlock()
		f(ip, peer)
		self.Mux.Lock()
	}
	self.Mux.Unlock()
}

func (self *PeersPool) Delete(ip string) {
	self.Mux.Lock()
	defer self.Mux.Unlock()
	delete(self.Peers, ip)

}

func (self *PeersPool) ConnIsExist(ip string) bool {
	if self.Exist(ip) {
		peer := self.Get(ip)
		return peer.Conn != nil
	} else {
		return false
	}
}

func (self *PeersPool) Sync() {
	res := Peers.Request(Join([][]byte{
		[]byte("sync peers"),
		[]byte{},
	}),
		func(blob []byte) bool {
			return true
		})
	self.AddNew(res)
	self.Clear()
}

func (self *PeersPool) NumSub() {
	self.Mux.Lock()
	defer self.Mux.Unlock()
	self.Num--
}

func (self *PeersPool) Add(peer *Peer) {
	self.Mux.Lock()
	defer self.Mux.Unlock()
	if _, ok := self.Peers[peer.IP()]; !ok {
		self.Peers[peer.IP()] = peer
	}
}

func (self *PeersPool) Encode() []byte {
	self.Mux.Lock()
	defer self.Mux.Unlock()
	var res [][]byte
	for _, p := range self.Peers {
		res = append(res, []byte(p.IP()))
	}
	return Join(res)
}

func (self *PeersPool) AddNew(blob []byte) {
	res := Split(blob)
	for _, p := range res {
		self.AddByIP(string(p))
	}
}

func (self *PeersPool) Get(ip string) *Peer {
	self.Mux.Lock()
	defer self.Mux.Unlock()
	return self.Peers[ip]
}

func (self *PeersPool) AddByIP(ip string) {
	self.Add(&Peer{
		RemoteAddr: ip,
	})
}

func NewPeer(conn net.Conn) *Peer {
	if conn == nil {
		return nil
	}
	peer := &Peer{
		RemoteAddr:   conn.RemoteAddr().String(),
		Conn:         conn,
		WriteChan:    make(chan []byte),
		ReadChan:     make(chan []byte),
		TmpChan:      make(chan []byte),
		RequestChan:  make(chan []byte),
		ResponseChan: make(chan []byte),
	}
	Peers.Mux.Lock()
	defer Peers.Mux.Unlock()
	Peers.Peers[peer.IP()] = peer
	return peer
}

func (self *Peer) Read() ([]byte, error) {
	conn := self.Conn
	if conn == nil {
		return nil, errconn
	}
	return DecodePrefixReader(conn)
}

func (self *Peer) Write(blob []byte) (int, error) {
	conn := self.Conn
	if conn == nil {
		return 0, errconn
	}
	return conn.Write(EncodePrefix(blob))
}

func (self *Peer) Close() {
	defer func() {
		recover()
	}()

	conn := self.Conn
	if conn != nil {
		conn.Close()
		self.Conn = nil
		close(self.WriteChan)
		close(self.ReadChan)
		close(self.TmpChan)
		close(self.ResponseChan)
		close(self.RequestChan)
		Peers.NumSub()
	}
}

func (self *PeersPool) ConnectToPeers() {
	for {
		self.Map(func(ip string, peer *Peer) {
			if self.Num > MaxPeers/2 {
				time.Sleep(300 * time.Second)
				return
			}
			if peer.Conn == nil {
				go self.Connect(ip)
			}

		})
		time.Sleep(10 * time.Second)
	}
}

func (self *PeersPool) Request(body []byte, check func(blob []byte) bool) []byte {
	defer func() {
		recover()
	}()
	ch := make(chan []byte)
	kill := make(chan int)
	self.Map(func(ip string, peer *Peer) {
		go func() {
			defer func() {
				recover()
			}()
			if peer.Conn != nil {
				select {
				case peer.RequestChan <- body:
				case kill <- 0:
				}
				select {
				case tmp := <-peer.ResponseChan:
					if check(tmp) {
						ch <- tmp
					}
				case kill <- 0:
				}
			}
		}()
		time.Sleep(Delay * time.Millisecond)

	})
	select {
	case tmp := <-ch:
		close(ch)
		close(kill)
		return tmp
	case <-time.After(TimeOut * time.Second):
		close(ch)
		close(kill)
		return nil
	}
}

func HandlePeer(peer *Peer) {
	if peer == nil || peer.Conn == nil {
		return
	}
	Peers.NumAdd()
	if Peers.Num > MaxPeers {
		peer.Close()
		return
	}
	ConnectionDeadline := time.Now().Add(time.Second * Deadline)
	peer.Conn.SetDeadline(ConnectionDeadline)

	go func() {
		defer func() {
			recover()
		}()
		var blob []byte
		var params [][]byte
		var err error
		for {
			blob, err = peer.Read()
			if err != nil {
				peer.Close()
				return
			}
			params = Split(blob)
			if len(params) != 2 {
				continue
			}
			switch string(params[0]) {
			case "request":
				peer.WriteChan <- Join([][]byte{
					[]byte("response"),
					HandleRequest(params[1]),
				})
			case "response":
				peer.TmpChan <- params[1]
			}
		}
	}()

	go func() {
		defer func() {
			recover()
		}()
		var err error
		var tmp []byte
		for {

			select {
			case tmp = <-peer.WriteChan:
				_, err = peer.Write(tmp)
				if err != nil {
					peer.Close()
					return
				}
			case <-time.After(ConnectionDeadline.Sub(time.Now())):
				peer.Close()
				return
			}
		}
	}()

	go func() {

		defer func() {
			recover()
		}()
		var tmp []byte
		for {
			select {
			case tmp = <-peer.RequestChan:
				peer.WriteChan <- Join([][]byte{
					[]byte("request"),
					tmp,
				})
				tmp = <-peer.TmpChan
				peer.ResponseChan <- tmp

			case <-time.After(ConnectionDeadline.Sub(time.Now())):
				return
			}
		}
	}()
}

func (self *Peer) IP() string {
	return IP(self.RemoteAddr)
}

func (self *PeersPool) Exist(key string) bool {
	self.Mux.Lock()
	defer self.Mux.Unlock()
	_, exist := self.Peers[key]
	return exist
}

func (self *Peer) Request(blob []byte) []byte {
	defer func() {
		recover()
	}()
	self.RequestChan <- blob
	return <-self.ResponseChan
}

func (self *PeersPool) Clear() {
	self.Mux.Lock()
	defer self.Mux.Unlock()
	if len(self.Peers) < MaxPeersLen {
		return
	}

	for k, v := range self.Peers {
		if v.Conn == nil && len(self.Peers) > MaxPeersLen {
			delete(self.Peers, k)
		}
	}
}
