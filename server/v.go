package main

import (
	"fmt"
	"github.com/juju/errors"
	"github.com/syndtr/goleveldb/leveldb"

	"github.com/dearcode/candy/meta"
	"github.com/dearcode/candy/util"
)

func main() {
	m := &meta.Message{}
	println(m)
	db, err := leveldb.OpenFile("/tmp/user.db", nil)
	if err != nil {
		e := errors.Trace(err)
		println(e.Error())
	}
	println(db)

	k := util.EncodeKey(int64(1), int64(2))
	fmt.Printf("%q\n", k)

}
