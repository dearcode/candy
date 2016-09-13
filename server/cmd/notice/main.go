package main

import (
	"flag"

	"github.com/dearcode/candy/server/notice"
	"github.com/dearcode/candy/server/util"
)

func main() {
	host := flag.String("p", "0.0.0.0:9003", "listen host")
	version := flag.Bool("v", false, "print version")
	flag.Parse()

	if *version {
		util.PrintVersion()
		return
	}

	s := notice.NewNotifer(*host)
	if err := s.Start(); err != nil {
		println(err.Error())
	}
}
