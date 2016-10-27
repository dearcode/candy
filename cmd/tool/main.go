package main

import (
	"flag"

	"github.com/fatih/color"
	"github.com/juju/errors"

	"github.com/dearcode/candy/util"
)

func getNewID(maddr, eaddrs string) {
	master, err := util.NewMasterClient(maddr, util.Split(eaddrs, ","))
	if err != nil {
		panic(errors.ErrorStack(err))
	}

	id, err := master.NewID()
	if err != nil {
		panic(errors.ErrorStack(err))
	}
	color.Green("new id:%d\n", id)
}

func getMasterAddr(eaddrs string) {
	etcd, err := util.NewEtcdClient(util.Split(eaddrs, ","))
	if err != nil {
		panic(errors.ErrorStack(err))
	}

	val, err := etcd.Get(util.EtcdMasterAddrKey)
	if err != nil {
		panic(errors.ErrorStack(err))
	}

	color.Green("master addr:%s\n", val)
}

func main() {
	m := flag.String("m", "0.0.0.0:9001", "master addr")
	e := flag.String("e", "", "etcd addrs, like `192.168.199.1:2379,192.168.199.2:2379,192.168.199.3:2379`")
	gnid := flag.Bool("id", false, "get new id from master")
	gma := flag.Bool("maddr", false, "get master addr from etcd")
	flag.Parse()

	switch {
	case *gnid:
		getNewID(*m, *e)
	case *gma:
		getMasterAddr(*e)
	}

}
