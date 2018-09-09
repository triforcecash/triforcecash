package core

import (
	"bytes"
	"github.com/triforcecash/triforcecash/core/sign"
	"regexp"
	"math/big"
	"strings"
)

func (self *Host) data() []byte {
	var buf bytes.Buffer
	buf.Write([]byte(self.Addr))
	buf.Write([]byte(self.Port))
	buf.Write([]byte(self.Prot))
	buf.Write(self.Pub)
	buf.Write(self.Nonce)
	return buf.Bytes()
}

func (self *Host) Check() bool {
	return sign.VerSign(self.data(), self.Proof, self.Pub)
}

func (self *Host) Sign() {
	self.Proof, _ = sign.GenSign(self.data(), Priv)
}

func SplitAddr(address string) (addr, port string) {
	tmp := strings.Split(address, ":")
	if len(tmp) == 1 {
		return tmp[0], ""
	}
	if len(tmp) == 2 {
		return tmp[0], ":" + tmp[1]
	}
	return "", ""
}

func AddHost(host *Host) {
	if !HostExist(host.Addr) && !IsIgnored(host.Addr) && CorrectAddress(host.Addr) && host.Check() {
		host.Karma = 0
		hostsmux.Lock()
		Hosts[host.Addr] = host
		hostsmux.Unlock()
	}
}

func AddHostAddr(addr string) {
	a, p := SplitAddr(addr)
	hostsmux.Lock()
	Hosts[a] = &Host{Addr: a, Port: p, Prot: protocol}
	hostsmux.Unlock()
}

func UpdateHost(host *Host) {
	if HostExist(host.Addr) {
		hostsmux.Lock()
		host.Karma = Hosts[host.Addr].Karma
		hostsmux.Unlock()
	}
	if !IsIgnored(host.Addr) && CorrectAddress(host.Addr) && host.Check() {
		hostsmux.Lock()
		Hosts[host.Addr] = host
		hostsmux.Unlock()
	}
}

func HostExist(addr string) bool {
	hostsmux.Lock()
	_, ok := Hosts[addr]
	hostsmux.Unlock()
	return ok
}

func CorrectAddress(addr string) bool {
	m, _ := regexp.MatchString(`(^[a-zA-Z0-9\.\:\-]+$)`, addr)
	return m
}
func IgnoreAddress(addr string) {
	hostsignoremux.Lock()
	HostsIgnore[addr] = 0
	hostsignoremux.Unlock()
	hostsmux.Lock()
	delete(Hosts, addr)
	hostsmux.Unlock()
}

func IsIgnored(addr string) bool {
	hostsignoremux.Lock()
	_, ignored := HostsIgnore[addr]
	hostsignoremux.Unlock()
	return ignored
}


func CalculateTotalRate()*big.Int{
	total:=new(big.Int).SetInt64(1)

	MapHosts(
		func(addr string,host *Host){
			total.Add(total,coef(append(host.Pub,host.Nonce...)))
		})
	return total
}

func CalculateChanceToCreateBlock(){
	total:=new(big.Float).SetInt(CalculateTotalRate())
	MapHosts(
		func(addr string, host *Host){
			rate:=new(big.Float).SetInt(coef(append(host.Pub,host.Nonce...)))
			host.Part,_=rate.Quo(rate,total).Float64()
		})
}