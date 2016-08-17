package main

import (
	"flag"

	"github.com/dearcode/candy/notice"
)

func main() {
	host := flag.String("p", "127.0.0.1:9003", "listen host")

	flag.Parse()

	s := notice.NewNotifer(*host)
	if err := s.Start(); err != nil {
		println(err.Error())
	}
}
