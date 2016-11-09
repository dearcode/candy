package main

import (
	"flag"

	"github.com/dearcode/candy/master"
	"github.com/dearcode/candy/util"
)

func main() {
	host := flag.String("p", "0.0.0.0:9001", "bind host")
	etcd := flag.String("e", "", "etcd addrs, like `192.168.199.1:2379,192.168.199.2:2379,192.168.199.3:2379`")
	version := flag.Bool("v", false, "print version")
	flag.Parse()

	if *version {
		util.PrintVersion()
		return
	}

	if _, err := master.NewMaster(*host, util.Split(*etcd, ",")); err != nil {
		println(err.Error())
	}
}
