package store

import (
	"encoding/json"
	"fmt"

	"github.com/syndtr/goleveldb/leveldb"

	"github.com/dearcode/candy/meta"
)

type MessageDB struct {
	dir string
	db  *leveldb.DB
}

func newMessageDB(dir string) *MessageDB {
	return &MessageDB{dir: dir}
}

func (u *MessageDB) start() error {
	db, err := leveldb.OpenFile(u.dir, nil)
	if err != nil {
		return err
	}
	u.db = db
	return nil
}

func (u *MessageDB) add(msg meta.Message) error {
	txn, err := u.db.OpenTransaction()
	if err != nil {
		return err
	}

	buf, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	if err = txn.Put([]byte(msg.ID), buf, nil); err != nil {
		return err
	}

	if err = txn.Commit(); err != nil {
		return err
	}

	return nil
}

