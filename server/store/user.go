package store

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"

	"github.com/juju/errors"
	"github.com/syndtr/goleveldb/leveldb"
	lu "github.com/syndtr/goleveldb/leveldb/util"

	"github.com/dearcode/candy/server/util"
	"github.com/dearcode/candy/server/util/log"
)

type account struct {
	ID       int64
	Name     string
	Password string
	NickName string
	Avatar   []byte
}

type userDB struct {
	root   string
	db     *leveldb.DB
	friend *friendDB
}

func newUserDB(root string) *userDB {
	return &userDB{root: root}
}

func (u *userDB) start() error {
	path := fmt.Sprintf("%s/%s", u.root, util.UserDBPath)
	db, err := leveldb.OpenFile(path, nil)
	if err != nil {
		return errors.Trace(err)
	}
	u.db = db
	u.friend = newFriendDB(u)
	return nil
}

func (u *userDB) register(user, passwd string, id int64) error {
	log.Debugf("user:%v passwd:%v", user, passwd)
	txn, err := u.db.OpenTransaction()
	if err != nil {
		txn.Discard()
		return errors.Trace(err)
	}

	log.Debugf("open transaction finished")
	v, err := txn.Get([]byte(user), nil)
	if err != nil && err != leveldb.ErrNotFound {
		txn.Discard()
		return errors.Trace(err)
	}

	log.Debugf("check whether the user exist")

	if len(v) != 0 {
		txn.Discard()
		return errors.Errorf("user:%s exist info:%s", user, string(v))
	}

	buf, err := json.Marshal(account{Name: user, Password: passwd, ID: id})
	if err != nil {
		txn.Discard()
		return errors.Trace(err)
	}

	if err = txn.Put([]byte(user), buf, nil); err != nil {
		txn.Discard()
		return errors.Trace(err)
	}

	if err = txn.Commit(); err != nil {
		txn.Discard()
		return errors.Trace(err)
	}

	log.Debugf("user:%s, passwd:%s, id:%d", user, passwd, id)

	return nil
}

func (u *userDB) findUser(user string) (int64, error) {
	v, err := u.db.Get([]byte(user), nil)
	if err != nil {
		return 0, errors.Trace(err)
	}

	var a account

	if err = json.Unmarshal(v, &a); err != nil {
		return 0, errors.Trace(err)
	}

	return a.ID, nil
}

func (u *userDB) auth(user, passwd string) (int64, error) {
	v, err := u.db.Get([]byte(user), nil)
	if err != nil {
		return 0, errors.Trace(err)
	}

	var a account

	if err = json.Unmarshal(v, &a); err != nil {
		return 0, errors.Trace(err)
	}

	if a.Password != passwd {
		return 0, errors.Errorf("invalid passwd:%s expect:%s", string(v))
	}

	return a.ID, nil
}

func (u *userDB) updateUserPassword(user, passwd string) error {
	txn, err := u.db.OpenTransaction()
	if err != nil {
		txn.Discard()
		return errors.Trace(err)
	}

	log.Debugf("open transaction finished")
	v, err := txn.Get([]byte(user), nil)
	if err != nil && err != leveldb.ErrNotFound {
		txn.Discard()
		return errors.Trace(err)
	}

	log.Debugf("check whether the user exist")
	if len(v) == 0 {
		txn.Discard()
		return errors.Errorf("user:%s not exist info:%s", user, string(v))
	}

	var a account
	if err = json.Unmarshal(v, &a); err != nil {
		txn.Discard()
		return errors.Trace(err)
	}

	if passwd != "" && a.Password != passwd {
		a.Password = passwd
	}

	buf, err := json.Marshal(a)
	if err != nil {
		txn.Discard()
		return errors.Trace(err)
	}

	if err = txn.Put([]byte(user), buf, nil); err != nil {
		txn.Discard()
		return errors.Trace(err)
	}

	if err = txn.Commit(); err != nil {
		txn.Discard()
		return errors.Trace(err)
	}

	return nil
}

func (u *userDB) updateUserInfo(user, nickName string, avatar []byte) (int64, error) {
	txn, err := u.db.OpenTransaction()
	if err != nil {
		txn.Discard()
		return -1, errors.Trace(err)
	}

	log.Debugf("open transaction finished")
	v, err := txn.Get([]byte(user), nil)
	if err != nil && err != leveldb.ErrNotFound {
		txn.Discard()
		return -1, errors.Trace(err)
	}

	log.Debugf("check whether the user exist")
	if len(v) == 0 {
		txn.Discard()
		return -1, errors.Errorf("user:%s not exist info:%s", user, string(v))
	}

	var a account
	var update bool
	var id int64
	if err = json.Unmarshal(v, &a); err != nil {
		txn.Discard()
		return -1, errors.Trace(err)
	}

	id = a.ID
	log.Debugf("compare nickName, old:%v new:%v", a.NickName, nickName)
	if nickName != "" && a.NickName != nickName {
		a.NickName = nickName
		update = true
	}

	log.Debugf("compare avatar")
	if avatar != nil && bytes.Compare(a.Avatar, avatar) != 0 {
		a.Avatar = avatar
		update = true
	}

	//no change
	if !update {
		txn.Discard()
		return id, nil
	}

	buf, err := json.Marshal(a)
	if err != nil {
		txn.Discard()
		return -1, errors.Trace(err)
	}

	if err = txn.Put([]byte(user), buf, nil); err != nil {
		txn.Discard()
		return -1, errors.Trace(err)
	}

	if err = txn.Commit(); err != nil {
		txn.Discard()
		return -1, errors.Trace(err)
	}

	log.Debugf("success")

	return id, nil
}

func (u *userDB) getUserInfo(user string) (*account, error) {
	v, err := u.db.Get([]byte(user), nil)
	if err != nil {
		return nil, errors.Trace(err)
	}

	var a account
	if err = json.Unmarshal(v, &a); err != nil {
		return nil, errors.Trace(err)
	}

	return &a, nil
}

func (u *userDB) addMessage(userID int64, msgID int64) error {
	key := UserMessageKey(userID, msgID)
	val := util.EncodeInt64(msgID)
	if err := u.db.Put(key, val, nil); err != nil {
		return errors.Trace(err)
	}

	lastKey := UserLastMessageKey(userID)
	var lastID int64

	v, err := u.db.Get(lastKey, nil)
	if err != nil {
		if err != leveldb.ErrNotFound {
			return errors.Trace(err)
		}
	} else {
		lastID = util.DecodeInt64(v)
	}

	if lastID > msgID {
		return nil
	}
	log.Debugf("user:%d, lastMessageID:%d", userID, msgID)

	return u.db.Put(UserLastMessageKey(userID), util.EncodeInt64(msgID), nil)
}

func (u *userDB) getLastMessageID(userID int64) (int64, error) {
	lastKey := UserLastMessageKey(userID)
	v, err := u.db.Get(lastKey, nil)
	if err != nil {
		return 0, errors.Trace(err)
	}
	return util.DecodeInt64(v), nil

}

func (u *userDB) getMessageIDs(userID int64, lastMsgID int64) ([]int64, error) {
	start := UserMessageKey(userID, lastMsgID)
	end := UserMessageKey(userID, math.MaxInt64)
	var ids []int64

	it := u.db.NewIterator(&lu.Range{Start: start, Limit: end}, nil)
	for it.Next(); it.Valid(); it.Next() {
		ids = append(ids, util.DecodeInt64(it.Value()))
	}

	return ids, nil
}
