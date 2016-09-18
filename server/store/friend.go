package store

import (
	"encoding/json"

	"github.com/juju/errors"
	"github.com/syndtr/goleveldb/leveldb"
	lu "github.com/syndtr/goleveldb/leveldb/util"
)

type friendRelation struct {
	ID      int64
	Confirm bool
}

type friendDB struct {
	*userDB
}

func newFriendDB(db *userDB) *friendDB {
	return &friendDB{userDB: db}
}

func (f *friendDB) add(uid, fid int64, confirm bool) error {
	var r friendRelation

	key := UserFriendKey(uid, fid)
	v, err := f.db.Get(key, nil)
	if err != nil {
		if confirm {
			//发确认消息，必须先有未验证的关系
			return errors.Annotatef(err, "uid:%d, fid:%d", uid, fid)
		}

		if err != leveldb.ErrNotFound {
			return errors.Trace(err)
		}
	} else {
		if err = json.Unmarshal(v, &r); err != nil {
			return errors.Trace(err)
		}
		if r.Confirm == confirm {
			return nil
		}

	}
	r = friendRelation{ID: fid, Confirm: confirm}

	buf, err := json.Marshal(&r)
	if err != nil {
		return errors.Trace(err)
	}

	return f.db.Put(key, buf, nil)
}

func (f *friendDB) get(uid int64) ([]int64, error) {
	var r friendRelation
	var ids []int64

	start, end := UserFriendRange(uid)
	it := f.db.NewIterator(&lu.Range{Start: start, Limit: end}, nil)

	for it.Next(); it.Valid(); it.Next() {
		if err := json.Unmarshal(it.Value(), &r); err != nil {
			return nil, errors.Trace(err)
		}

		if r.Confirm {
			ids = append(ids, r.ID)
		}
	}

	return ids, nil
}

func (f *friendDB) exist(uid, fid int64) error {
	v, err := f.db.Get(UserFriendKey(uid, fid), nil)
	if err != nil {
		return errors.Trace(err)
	}

	var r friendRelation
	if err = json.Unmarshal(v, &r); err != nil {
		return errors.Trace(err)
	}

	if !r.Confirm {
		return leveldb.ErrNotFound
	}

	return nil
}
