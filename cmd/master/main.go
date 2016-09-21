package main

import (
	"flag"

	"github.com/dearcode/candy/master"
	"github.com/dearcode/candy/util"
)

func main() {
	host := flag.String("p", "0.0.0.0:9001", "listen host")
	version := flag.Bool("v", false, "print version")
	flag.Parse()

	if *version {
		util.PrintVersion()
		return
	}

	s := master.NewMasterServer(*host)
	if err := s.Start(); err != nil {
		println(err.Error())
	}
}
