package main

import (
	"flag"

	"github.com/dearcode/candy/server/gate"
	"github.com/dearcode/candy/server/util"
)

func main() {
	host := flag.String("p", "0.0.0.0:9000", "listen host")
	master := flag.String("m", "0.0.0.0:9001", "master host")
	store := flag.String("s", "0.0.0.0:9004", "store host")
	notice := flag.String("n", "0.0.0.0:9003", "notice host")
	version := flag.Bool("v", false, "print version")
	flag.Parse()

	if *version {
		util.PrintVersion()
		return
	}

	s := gate.NewGate()
	if err := s.Start(*host, *notice, *master, *store); err != nil {
		println(err.Error())
	}

}
