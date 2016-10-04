package store

import (
	"encoding/json"

	"github.com/juju/errors"
	"github.com/syndtr/goleveldb/leveldb"
	lu "github.com/syndtr/goleveldb/leveldb/util"

	"github.com/dearcode/candy/meta"
	"github.com/dearcode/candy/util/log"
)

type friendRelation struct {
	ID    int64
	State meta.Relation
	Msg   string
}

type friendDB struct {
	*userDB
}

func newFriendDB(db *userDB) *friendDB {
	return &friendDB{userDB: db}
}

// 修改好友关系
func (f *friendDB) set(uid, fid int64, state meta.Relation, msg string) error {
	key := UserFriendKey(uid, fid)

	r := friendRelation{ID: fid, State: state, Msg: msg}
	buf, err := json.Marshal(&r)
	if err != nil {
		return errors.Trace(err)
	}

	if err = f.db.Put(key, buf, nil); err != nil {
		return errors.Trace(err)
	}

	return nil
}

func (f *friendDB) confirm(uid, fid int64) error {
	key := UserFriendKey(uid, fid)
	v, err := f.db.Get(key, nil)
	if err != nil {
		return errors.Trace(err)
	}
	var r friendRelation
	if err = json.Unmarshal(v, &r); err != nil {
		return errors.Trace(err)
	}

	if r.State != meta.Relation_ADD {
		log.Infof("%d friend:%d state:%v", uid, fid, r.State)
		return nil
	}

	r.State = meta.Relation_CONFIRM

	buf, err := json.Marshal(&r)
	if err != nil {
		return errors.Trace(err)
	}

	if err = f.db.Put(key, buf, nil); err != nil {
		return errors.Trace(err)
	}

	return nil
}

// 删除好友关系
func (f *friendDB) remove(uid, fid int64) error {
	key := UserFriendKey(uid, fid)
	return f.db.Delete(key, nil)
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

		if r.State == meta.Relation_CONFIRM {
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

	if r.State != meta.Relation_CONFIRM {
		return leveldb.ErrNotFound
	}

	return nil
}
