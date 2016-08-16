package main

import (
	"flag"

	"github.com/dearcode/candy/gate"
)

func main() {
	host := flag.String("p", "127.0.0.1:9000", "listen host")
	flag.Parse()

	s := gate.NewGateServer(*host)
	s.Start()
}
