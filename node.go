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
	checkdepth := flag.Int("checkdepth", 1000, "For a stronger check, you should set 10000.")
	seed := flag.String("seed", "", "Seed (password from your account")
	lobby := flag.String("lobby", "185.234.15.72:8075", "Lobby node")
	fullnode := flag.Bool("fullnode", false, "Will be able fullnode features")
	flag.Parse()

	if *seed != "" {
		core.SetSeed([]byte(*seed))
		fmt.Printf("Your address: %x\n", core.Addr(core.Pub))
		core.Mineblocks = true
		core.Minecpu = true
	}
	core.FullNode = *fullnode
	core.Checkdepth = *checkdepth
	core.Port = fmt.Sprint(":", *port)
	core.PublicIp = *hostname
	if *hostname == "127.0.0.1" {
		core.ClientOnly = true
	}

	core.Start()
	core.AddHostAddr(*lobby)
	defer core.Stop()
	fmt.Println("Press Сtr+C to stop")

	var wg sync.WaitGroup
	wg.Add(1)
	wg.Wait()
}
