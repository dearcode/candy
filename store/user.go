package store

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"strings"

	"github.com/juju/errors"
	"github.com/syndtr/goleveldb/leveldb"
	lu "github.com/syndtr/goleveldb/leveldb/util"

	"github.com/dearcode/candy/util"
	"github.com/dearcode/candy/util/log"
)

const (
	//模糊查询用户最大数量
	maxFindUserCount = 20
)

type account struct {
	ID       int64
	Name     string
	Password string
	NickName string
	Avatar   string
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
	log.Debugf("path:%v", path)
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

	//创建用户的同时增加userid+@+username的编码key，提供通过userid查找用户数据
	buf, err = json.Marshal(user)
	if err != nil {
		txn.Discard()
		return errors.Trace(err)
	}

	if err = txn.Put(UserUnionKey(id), buf, nil); err != nil {
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

func (u *userDB) findUser(user string) ([]string, error) {
	var users []string
	count := 0

	it := u.db.NewIterator(&lu.Range{Start: []byte(user)}, nil)
	for it.Next(); it.Valid(); it.Next() {
		var a account
		if err := json.Unmarshal(it.Value(), &a); err != nil {
			return nil, errors.Trace(err)
		}

		//对比如果查找到的用户名称不包含所查找的串就跳过
		log.Debugf("a.Name:%v user:%v", a.Name, user)
		if !strings.Contains(a.Name, user) {
			continue
		}

		users = append(users, a.Name)
		count = count + 1
		if count >= maxFindUserCount {
			break
		}
	}

	return users, nil
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
		return 0, errors.Errorf("invalid passwd:%s", string(passwd))
	}

	return a.ID, nil
}

// resetUserPassword 后台接口，直接重置用户密码
func (u *userDB) resetUserPassword(name, passwd string) error {
	log.Debugf("begin user:%s reset passwd:%s", name, passwd)
	txn, err := u.db.OpenTransaction()
	if err != nil {
		txn.Discard()
		return errors.Trace(err)
	}

	buf, err := txn.Get([]byte(name), nil)
	if err != nil {
		txn.Discard()
		return errors.Trace(err)
	}

	var a account
	if err = json.Unmarshal(buf, &a); err != nil {
		txn.Discard()
		return errors.Trace(err)
	}

	a.Password = passwd

	if buf, err = json.Marshal(a); err != nil {
		txn.Discard()
		return errors.Trace(err)
	}

	if err = txn.Put([]byte(name), buf, nil); err != nil {
		txn.Discard()
		return errors.Trace(err)
	}

	if err = txn.Commit(); err != nil {
		txn.Discard()
		return errors.Trace(err)
	}
	log.Infof("end user:%s new passwd:%s", name, passwd)

	return nil
}

// updateUserPassword 用户自己修改密码，需要原密码
func (u *userDB) updateUserPassword(id int64, name, oldPasswd, newPasswd string) error {
	log.Debugf("begin %d user:%s change passwd from:%s to:%s", id, name, oldPasswd, newPasswd)
	txn, err := u.db.OpenTransaction()
	if err != nil {
		txn.Discard()
		return errors.Trace(err)
	}

	buf, err := txn.Get([]byte(name), nil)
	if err != nil {
		txn.Discard()
		return errors.Trace(err)
	}

	var a account
	if err = json.Unmarshal(buf, &a); err != nil {
		txn.Discard()
		return errors.Trace(err)
	}

	if oldPasswd != a.Password {
		txn.Discard()
		return errors.Trace(ErrInvalidOperator)
	}

	a.Password = newPasswd

	if buf, err = json.Marshal(a); err != nil {
		txn.Discard()
		return errors.Trace(err)
	}

	if err = txn.Put([]byte(name), buf, nil); err != nil {
		txn.Discard()
		return errors.Trace(err)
	}

	if err = txn.Commit(); err != nil {
		txn.Discard()
		return errors.Trace(err)
	}

	log.Infof("end %d name:%s new passwd:%s", id, name, newPasswd)
	return nil
}

//updateUserInfo 如果nickname没改，avatar也没改，就不要调用这个函数了
func (u *userDB) updateUserInfo(id int64, name, nickName string, avatar string) error {
	txn, err := u.db.OpenTransaction()
	if err != nil {
		return errors.Trace(err)
	}

	log.Debugf("%d open transaction finished", id)
	buf, err := txn.Get([]byte(name), nil)
	if err != nil {
		txn.Discard()
		return errors.Trace(err)
	}

	var a account
	if err = json.Unmarshal(buf, &a); err != nil {
		txn.Discard()
		return errors.Trace(err)
	}

	if a.ID != id {
		txn.Discard()
		return errors.Trace(ErrInvalidOperator)
	}

	if nickName != "" && a.NickName != nickName {
		log.Debugf("%d nickName old:%v new:%v", id, a.NickName, nickName)
		a.NickName = nickName
	}

	if avatar != "" && a.Avatar != avatar {
		log.Debugf("%d avatar old:%v new:%v", id, a.Avatar, avatar)
		a.Avatar = avatar
	}

	if buf, err = json.Marshal(a); err != nil {
		txn.Discard()
		return errors.Trace(err)
	}

	if err = txn.Put([]byte(name), buf, nil); err != nil {
		txn.Discard()
		return errors.Trace(err)
	}

	if err = txn.Commit(); err != nil {
		txn.Discard()
		return errors.Trace(err)
	}

	log.Debugf("%d update info success", id)
	return nil
}

func (u *userDB) getUserInfoByName(name string) (*account, error) {
	v, err := u.db.Get([]byte(name), nil)
	if err != nil {
		return nil, errors.Trace(err)
	}

	var a account
	if err = json.Unmarshal(v, &a); err != nil {
		return nil, errors.Trace(err)
	}

	return &a, nil
}

func (u *userDB) getUserInfoByID(id int64) (*account, error) {
	v, err := u.db.Get(UserUnionKey(id), nil)
	if err != nil {
		return nil, errors.Trace(err)
	}

	var name string
	if err = json.Unmarshal(v, &name); err != nil {
		return nil, errors.Trace(err)
	}

	log.Debugf("get username by userid, name:%v", name)

	a, err := u.getUserInfoByName(name)
	if err != nil {
		return nil, errors.Trace(err)
	}

	return a, nil
}

func (u *userDB) addMessage(userID int64, msgID int64) (int64, error) {
	key := UserMessageKey(userID, msgID)
	val := util.EncodeInt64(msgID)
	if err := u.db.Put(key, val, nil); err != nil {
		return 0, errors.Trace(err)
	}

	before := int64(0)
	v, err := u.db.Get(UserLastMessageKey(userID), nil)
	if err == nil {
		before = util.DecodeInt64(v)
	} else if err != leveldb.ErrNotFound {
		return 0, errors.Trace(err)
	}

	log.Debugf("user:%d, before:%d, new:%d", userID, before, msgID)

	return before, u.db.Put(UserLastMessageKey(userID), util.EncodeInt64(msgID), nil)
}

func (u *userDB) getLastID(userID int64) (int64, error) {
	v, err := u.db.Get(UserLastMessageKey(userID), nil)
	if err != nil {
		return 0, errors.Trace(err)
	}
	return util.DecodeInt64(v), nil
}

func (u *userDB) getMessage(userID int64, reverse bool, id int64) ([]int64, error) {
	start := UserMessageKey(userID, id)
	end := UserMessageKey(userID, math.MaxInt64)
	if reverse {
		end = UserMessageKey(userID, 0)
	}
	var ids []int64

	it := u.db.NewIterator(nil, nil)

	for ok := it.Seek(start); ok; {
		ids = append(ids, util.DecodeInt64(it.Value()))
		if reverse {
			ok = it.Prev() && bytes.Compare(end, it.Key()) <= 0
		} else {
			ok = it.Next() && bytes.Compare(end, it.Key()) >= 0
		}
	}

	it.Release()

	return ids, nil
}

func (u *userDB) addGroup(userID int64, groupID int64) error {
	key := UserGroupKey(userID, groupID)
	val := util.EncodeInt64(groupID)
	return errors.Trace(u.db.Put(key, val, nil))
}

func (u *userDB) delGroup(userID int64, groupID int64) error {
	key := UserGroupKey(userID, groupID)
	return errors.Trace(u.db.Delete(key, nil))
}

func (u *userDB) getGroups(userID int64) ([]int64, error) {
	start, end := UserGroupRange(userID)

	var ids []int64
	it := u.db.NewIterator(nil, nil)
	for ok := it.Seek(start); ok; {
		ids = append(ids, util.DecodeInt64(it.Value()))

		ok = it.Next() && bytes.Compare(end, it.Key()) >= 0
	}

	it.Release()

	return ids, nil
}
