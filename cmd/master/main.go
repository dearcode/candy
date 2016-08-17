package main

import (
	"flag"

	"github.com/dearcode/candy/master"
)

func main() {
	host := flag.String("p", "127.0.0.1:9000", "listen host")
	flag.Parse()

	s := master.NewMasterServer(*host)
	s.Start()
}
