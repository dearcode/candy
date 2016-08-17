package main

import (
	"flag"

	"github.com/dearcode/candy/gate"
)

func main() {
	host := flag.String("p", "127.0.0.1:9000", "listen host")
	master := flag.String("p", "127.0.0.1:9001", "master host")
	store := flag.String("p", "127.0.0.1:9004", "store host")

	flag.Parse()

	s := gate.NewGate(*host, *master, *store)
	if err := s.Start(); err != nil {
		println(err.Error())
	}

}
