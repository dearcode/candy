package master

import (
	"strconv"
	"time"

	"github.com/dearcode/candy/util/log"
)

type mstore struct {
	v string
}

func newMstore() *mstore {
	return &mstore{}
}

func (m *mstore) Get(k string) (string, error) {
	return strconv.FormatInt(time.Now().Unix(), 10), nil
}

func (m *mstore) CAS(_, _, _, v string) error {
	m.v = v
	log.Debugf("current last:%s", v)
	return nil
}
