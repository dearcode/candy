package main

import (
	"flag"

	"github.com/dearcode/candy/server/gate"
)

func main() {
	host := flag.String("p", "127.0.0.1:9000", "listen host")
	master := flag.String("m", "127.0.0.1:9001", "master host")
	store := flag.String("s", "127.0.0.1:9004", "store host")

	flag.Parse()

	s := gate.NewGate(*host, *master, *store)
	if err := s.Start(); err != nil {
		println(err.Error())
	}

}
