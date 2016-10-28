package main

import (
	"flag"

	"github.com/dearcode/candy/notice"
	"github.com/dearcode/candy/util"
)

func main() {
	host := flag.String("p", "0.0.0.0:9003", "listen host")
	version := flag.Bool("v", false, "print version")
	flag.Parse()

	if *version {
		util.PrintVersion()
		return
	}

	_, err := notice.NewNotifer(*host)
	if err != nil {
		panic(err.Error())
	}
}
