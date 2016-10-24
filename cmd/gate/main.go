package main

import (
	"flag"

	"github.com/dearcode/candy/gate"
	"github.com/dearcode/candy/util"
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

	if _, err := gate.NewGate(*host, *master, *notice, *store); err != nil {
		println(err.Error())
	}

}
