package main

import (
	"flag"
	"time"

	"github.com/fatih/color"
	"github.com/juju/errors"

	"github.com/dearcode/candy/meta"
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

func sendMessageToUserID(id int64, addr string) {
	n, err := util.NewNotiferClient(addr)
	if err != nil {
		panic(errors.ErrorStack(err))
	}

	msg := meta.PushMessage{Msg: meta.Message{Body: time.Now().Format(time.ANSIC)}}
	pid := meta.PushID{User: id, Before: time.Now().UnixNano()}
	if err = n.Push(msg, pid); err != nil {
		panic(errors.ErrorStack(err))
	}

	color.Green("send to:%d message:%+v ok\n", id, msg)
}

func sendMessageToUserName(name, addr string) {
	//TODO

}

func main() {
	m := flag.String("m", "0.0.0.0:9001", "master addr")
	n := flag.String("n", "0.0.0.0:9003", "notice server addr")
	e := flag.String("e", "", "etcd addrs, like `192.168.199.1:2379,192.168.199.2:2379,192.168.199.3:2379`")
	gnid := flag.Bool("id", false, "get new id from master")
	gma := flag.Bool("maddr", false, "get master addr from etcd")
	stid := flag.Int64("stid", 0, "send message to user ID")
	stn := flag.String("stn", "", "send message to user Name")
	flag.Parse()

	switch {
	case *gnid:
		getNewID(*m, *e)
	case *gma:
		getMasterAddr(*e)
	case *stid != 0:
		sendMessageToUserID(*stid, *n)
	case *stn != "":
		sendMessageToUserName(*stn, *n)
	}

}
