package main

import (
	"fmt"

	"github.com/juju/errors"
	"github.com/syndtr/goleveldb/leveldb"

	"github.com/dearcode/candy/meta"
	"github.com/dearcode/candy/util"
	"github.com/dearcode/candy/util/log"
)

func main() {
	msg := meta.Message{}
	fmt.Printf("msg:%+v\n", msg)

	fmt.Printf("%s\n", util.EncodeInt64(123))
	log.Debug("xxx")

	err := fmt.Errorf("abc")
	err = errors.Trace(err)
	println(err)

	db := leveldb.DB{}
	fmt.Printf("%v\n", db)
}
