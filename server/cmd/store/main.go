package main

import (
	"flag"

	"github.com/dearcode/candy/server/store"
)

func main() {
	host := flag.String("p", "127.0.0.1:9004", "listen host")
	notice := flag.String("n", "127.0.0.1:9003", "notice host")
	path := flag.String("d", "/tmp/candy.db", "db path")

	flag.Parse()

	s := store.NewStore(*host, *notice, *path)
	if err := s.Start(); err != nil {
		println(err.Error())
	}
}
