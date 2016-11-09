package main

import (
	"flag"

	"github.com/dearcode/candy/notice"
	"github.com/dearcode/candy/util"
)

func main() {
	host := flag.String("h", "0.0.0.0:9003", "bind host")
	master := flag.String("m", "0.0.0.0:9001", "master host")
	etcd := flag.String("e", "", "etcd addrs, like `192.168.199.1:2379,192.168.199.2:2379,192.168.199.3:2379`")
	version := flag.Bool("v", false, "print version")
	flag.Parse()

	if *version {
		util.PrintVersion()
		return
	}

	if _, err := notice.NewServer(*host, *master, *etcd); err != nil {
		panic(err.Error())
	}
}
