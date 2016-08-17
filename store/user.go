package store

import (
	"encoding/json"
	"fmt"

	"github.com/syndtr/goleveldb/leveldb"
)

type UserInfo struct {
	ID       int64
	Name     string
	Password string
}

type userDB struct {
	dir string
	db  *leveldb.DB
}

func newUserDB(dir string) *userDB {
	return &userDB{dir: dir}
}

func (u *userDB) start() error {
	db, err := leveldb.OpenFile(u.dir, nil)
	if err != nil {
		return err
	}
	u.db = db
	return nil
}

func (u *userDB) register(user, passwd string, id int64) error {
	txn, err := u.db.OpenTransaction()
	if err != nil {
		return err
	}

	v, err := txn.Get([]byte(user), nil)
	if err != nil && err != leveldb.ErrNotFound {
		return err
	}

	if len(v) != 0 {
		return fmt.Errorf("user:%s exist info:%s", user, string(v))
	}

	buf, err := json.Marshal(UserInfo{Name: user, Password: passwd, ID: id})
	if err != nil {
		return err
	}

	if err = txn.Put([]byte(user), buf, nil); err != nil {
		return err
	}

	if err = txn.Commit(); err != nil {
		return err
	}

	return nil
}

func (u *userDB) auth(user, passwd string) (int64, error) {
	v, err := u.db.Get([]byte(user), nil)
	if err != nil {
		return 0, err
	}

	var info UserInfo

	if err = json.Unmarshal(v, &info); err != nil {
		return 0, err
	}

	if info.Password != passwd {
		return 0, fmt.Errorf("invalid passwd:%s expect:%s", string(v))
	}

	return info.ID, nil
}
