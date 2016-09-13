package main

import (
	"flag"

	"github.com/dearcode/candy/server/store"
	"github.com/dearcode/candy/server/util"
)

func main() {
	host := flag.String("p", "0.0.0.0:9004", "listen host")
	notice := flag.String("n", "0.0.0.0:9003", "notice host")
	path := flag.String("d", "/tmp/candy.db", "db path")
	version := flag.Bool("v", false, "print version")
	flag.Parse()

	if *version {
		util.PrintVersion()
		return
	}

	s := store.NewStore(*host, *notice, *path)
	if err := s.Start(); err != nil {
		println(err.Error())
	}
}
