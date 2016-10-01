package store

import (
	"encoding/json"
	"fmt"
	"math"

	"github.com/juju/errors"
	"github.com/syndtr/goleveldb/leveldb"
	lu "github.com/syndtr/goleveldb/leveldb/util"

	"github.com/dearcode/candy/meta"
	"github.com/dearcode/candy/util"
	"github.com/dearcode/candy/util/log"
)

type groupDB struct {
	root string
	db   *leveldb.DB
}

func newGroupDB(dir string) *groupDB {
	return &groupDB{root: dir}
}

func (g *groupDB) start() error {
	path := fmt.Sprintf("%s/%s", g.root, util.GroupDBPath)
	log.Debugf("path:%v", path)
	db, err := leveldb.OpenFile(path, nil)
	if err != nil {
		return errors.Trace(err)
	}

	g.db = db

	return nil
}

// 创建一个新的组
func (g *groupDB) newGroup(group meta.GroupInfo) error {
	buf, err := json.Marshal(group)
	if err != nil {
		return errors.Trace(err)
	}

	return g.db.Put(util.EncodeInt64(group.ID), buf, nil)
}

// 向一个组中添加用户
func (g *groupDB) addUser(gid int64, ids ...int64) error {
	for _, uid := range ids {
		key := util.EncodeKey(gid, uid)
		if err := g.db.Put(key, []byte(""), nil); err != nil {
			return errors.Trace(err)
		}
	}
	return nil
}

// 获取组信息及所有用户ID
func (g *groupDB) get(gid int64) (meta.GroupInfo, error) {
	val, err := g.db.Get(util.EncodeInt64(gid), nil)
	if err != nil {
		return meta.GroupInfo{}, errors.Trace(err)
	}

	var group meta.GroupInfo

	if err = json.Unmarshal(val, &group); err != nil {
		return meta.GroupInfo{}, errors.Trace(err)
	}

	start := util.EncodeKey(gid, 0)
	end := util.EncodeKey(gid, math.MaxInt64)

	it := g.db.NewIterator(&lu.Range{Start: start, Limit: end}, nil)
	for it.Next(); it.Valid(); it.Next() {
		group.Users = append(group.Users, util.DecodeInt64(it.Key()[8:]))
	}

	return group, nil
}

func (g *groupDB) exist(gid, uid int64) error {
	_, err := g.db.Get(util.EncodeKey(gid, uid), nil)
	return err
}
