package main

import (
	"flag"
	"fmt"
	"github.com/triforcecash/triforcecash/core"
)

func main() {

	port := flag.Int("port", 0, "Http server port")
	checkdepth := flag.Int("checkdepth", 0, "For a stronger check, you should set 10000.")
	seed := flag.String("seed", "", "Seed (password from your account")
	lobby := flag.String("lobby", "", "Lobby node")
	flag.Parse()

	if *seed != "" {
		core.SetSeed([]byte(*seed))
		fmt.Printf("Your address: %x\n", core.Addr(core.Pub))
		core.Mineblocks = true
		core.Minecpu = true
	}

	if *checkdepth != 0 {
		core.Checkdepth = *checkdepth
	}

	if *port != 0 {
		core.PortHTTP = fmt.Sprint(":", *port)

	}

	if *lobby != "" {
		core.Lobby = *lobby
	}

	defer core.Stop()
	core.Start()
}
