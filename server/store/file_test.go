package store

import (
	"bytes"
	"testing"

	"github.com/dearcode/candy/server/util"
)

var (
	f *fileDB
)

func init() {
	f = newFileDB("/tmp/filedb")
	if err := f.start(); err != nil {
		panic(err.Error())
	}
}

func TestFileDB(t *testing.T) {
	dat := []byte("xxxxxxxzzzzzzzzzzzAAAAAAABBBBBBBBB")
	key := util.MD5(dat)
	if err := f.add(key, dat); err != nil {
		t.Fatalf("add key error:%s", err.Error())
	}
	ok, err := f.exist(key)
	if err != nil {
		t.Fatalf("exist error:%s", err.Error())
	}

	if !ok {
		t.Fatalf("expect exists")
	}

	val, err := f.get(key)
	if err != nil {
		t.Fatalf("get key error:%s", err.Error())
	}
	if !bytes.Equal(val, dat) {
		t.Fatalf("src val:%q, get val:%q", dat, val)
	}
}
