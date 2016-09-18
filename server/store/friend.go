package store

import (
	"encoding/json"

	"github.com/juju/errors"
	"github.com/syndtr/goleveldb/leveldb"
	lu "github.com/syndtr/goleveldb/leveldb/util"

	"github.com/dearcode/candy/server/meta"
)

type friendRelation struct {
	ID    int64
	State meta.FriendRelation
}

type friendDB struct {
	*userDB
}

func newFriendDB(db *userDB) *friendDB {
	return &friendDB{userDB: db}
}

// 添加好友，返回当前状态，state 0:没关系, 1:我要添加对方为好友, 2:对方请求添加我为好友, 3:当前我们都已确认成为好友了
func (f *friendDB) add(uid, fid int64, state meta.FriendRelation) (meta.FriendRelation, error) {
	r := friendRelation{ID: fid, State: state}

	key := UserFriendKey(uid, fid)
	v, err := f.db.Get(key, nil)
	if err != nil && err != leveldb.ErrNotFound {
		return meta.FriendRelation_None, errors.Trace(err)
	}

	if len(v) > 0 {
		if err = json.Unmarshal(v, &r); err != nil {
			return meta.FriendRelation_None, errors.Trace(err)
		}
	}

	r.State |= state

	buf, err := json.Marshal(&r)
	if err != nil {
		return meta.FriendRelation_None, errors.Trace(err)
	}

	if err = f.db.Put(key, buf, nil); err != nil {
		return meta.FriendRelation_None, errors.Trace(err)
	}

	return r.State, nil
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

		if r.State == meta.FriendRelation_Confirm {
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

	if r.State != 3 {
		return leveldb.ErrNotFound
	}

	return nil
}
