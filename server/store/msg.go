package store

import (
	"encoding/json"
	"fmt"

	"github.com/juju/errors"
	"github.com/syndtr/goleveldb/leveldb"

	"github.com/dearcode/candy/server/meta"
	"github.com/dearcode/candy/server/util"
)

type messageDB struct {
	root string
	db   *leveldb.DB
}

func newMessageDB(dir string) *messageDB {
	return &messageDB{root: dir}
}

func (m *messageDB) start() error {
	path := fmt.Sprintf("%s/%s", m.root, util.MessageDBPath)
	db, err := leveldb.OpenFile(path, nil)
	if err != nil {
		return err
	}
	m.db = db
	return nil
}

func (m *messageDB) add(msg meta.Message) error {
	buf, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	return m.db.Put(util.EncodeInt64(msg.ID), buf, nil)
}

func (m *messageDB) message(ids ...int64) ([]meta.Message, error) {
	var mss []meta.Message
	var msg meta.Message

	for _, id := range ids {
		v, err := m.db.Get(util.EncodeInt64(id), nil)
		if err != nil {
			return nil, errors.Trace(err)
		}

		if err = json.Unmarshal(v, &msg); err != nil {
			return nil, errors.Trace(err)
		}

		mss = append(mss, msg)
	}

	return mss, nil
}
