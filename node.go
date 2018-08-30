package main

import (
	"flag"
	"fmt"
	"github.com/triforcecash/triforcecash/core"
	"sync"
)

func main() {

	port := flag.Int("port", 8075, "Port")
	hostname := flag.String("host", "127.0.0.1", "Public ip")
	seed := flag.String("seed", "", "Seed (password from your account")
	lobby := flag.String("lobby", "185.234.15.72:8075", "Lobby node")
	clientonly := flag.Bool("client", false, "You will not be a host")
	flag.Parse()

	if *seed != "" {
		core.SetSeed([]byte(*seed))
		fmt.Printf("Your address: %x\n", core.Addr(core.Pub))
		core.Mineblocks = true
		core.Minecpu = true
	}

	core.Port = fmt.Sprint(":", *port)
	core.PublicIp = *hostname
	core.ClientOnly = *clientonly
	core.Start()
	core.AddHostAddr(*lobby)
	defer core.Stop()
	fmt.Println("Press Сtr+C to stop")

	var wg sync.WaitGroup
	wg.Add(1)
	wg.Wait()
}
