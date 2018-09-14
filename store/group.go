package store

import (
	"encoding/json"
	"fmt"
	"math"
	"time"

	"github.com/juju/errors"
	"github.com/syndtr/goleveldb/leveldb"
	lu "github.com/syndtr/goleveldb/leveldb/util"

	"github.com/dearcode/candy/meta"
	"github.com/dearcode/candy/util"
	"github.com/dearcode/candy/util/log"
)

var (
	// ErrInvalidOperator 当前用户没有权限进行这项操作.
	ErrInvalidOperator = errors.New("invalid operator")
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
func (g *groupDB) newGroup(gid, uid int64, name string) error {
	buf, err := json.Marshal(meta.GroupInfo{ID: gid, Owner: uid, Name: name, Active: true})
	if err != nil {
		return errors.Trace(err)
	}

	return g.db.Put(util.EncodeInt64(gid), buf, nil)
}

func (g *groupDB) delGroup(gid, uid int64) error {
	key := util.EncodeInt64(gid)
	val, err := g.db.Get(key, nil)
	if err != nil {
		return errors.Trace(err)
	}
	var group meta.GroupInfo
	if err = json.Unmarshal(val, &group); err != nil {
		return errors.Trace(err)
	}

	if group.Owner != uid {
		return errors.Trace(ErrInvalidOperator)
	}

	group.Active = false

	//TODO 延迟清理
	log.Infof("DeleteGroup:%d, User:%d", gid, uid)

	return errors.Trace(g.db.Delete(key, nil))
}

/*
func (g *groupDB) addAdmin(gid int64, ids ...int64) error {
	for _, uid := range ids {
		key := GroupAdminKey(gid, uid)
		if err := g.db.Put(key, []byte(""), nil); err != nil {
			return errors.Trace(err)
		}
	}
	return nil
}
*/

func (g *groupDB) getAdmins(gid int64) []int64 {
	var users []int64

	start := GroupAdminKey(gid, 0)
	end := GroupAdminKey(gid, math.MaxInt64)

	it := g.db.NewIterator(&lu.Range{Start: start, Limit: end}, nil)
	for it.Next(); it.Valid(); it.Next() {
		_, _, user := DecodeGroupKey(it.Key())
		users = append(users, user)
	}
	it.Release()

	return users
}

func (g *groupDB) delAdmin(gid int64, ids ...int64) error {
	for _, uid := range ids {
		if err := g.db.Delete(GroupAdminKey(gid, uid), nil); err != nil {
			return errors.Trace(err)
		}
	}
	return nil
}

// 向一个组中添加用户, 管理员和群主也要添加到成员中
func (g *groupDB) addMember(gid int64, ids ...int64) error {
	for _, uid := range ids {
		if err := g.db.Put(GroupMemberKey(gid, uid), []byte(""), nil); err != nil {
			return errors.Trace(err)
		}
	}
	return nil
}

func (g *groupDB) delMember(gid int64, ids ...int64) error {
	for _, uid := range ids {
		if err := g.db.Delete(GroupMemberKey(gid, uid), nil); err != nil {
			return errors.Trace(err)
		}
	}
	return nil
}

func (g *groupDB) getMember(gid int64) []int64 {
	var users []int64

	start := GroupMemberKey(gid, 0)
	end := GroupMemberKey(gid, math.MaxInt64)

	it := g.db.NewIterator(&lu.Range{Start: start, Limit: end}, nil)
	for it.Next(); it.Valid(); it.Next() {
		_, _, user := DecodeGroupKey(it.Key())
		users = append(users, user)
	}
	it.Release()

	return users
}

// getGroupHeader 获取群基本信息
func (g *groupDB) getGroupBase(gid int64) (meta.GroupInfo, error) {
	var group meta.GroupInfo

	val, err := g.db.Get(util.EncodeInt64(gid), nil)
	if err != nil {
		return group, errors.Trace(err)
	}

	err = json.Unmarshal(val, &group)

	return group, errors.Trace(err)
}

// 获取组信息及所有用户ID
func (g *groupDB) getGroup(gid int64) (meta.GroupInfo, error) {
	group, err := g.getGroupBase(gid)
	if err != nil {
		return group, errors.Trace(err)
	}

	group.Member = g.getMember(gid)
	group.Admins = g.getAdmins(gid)

	return group, nil
}

func (g *groupDB) exit(gid, uid int64) error {
	group, err := g.getGroupBase(gid)
	if err != nil {
		return errors.Trace(err)
	}

	//群主不能退出，只能解散群
	if group.Owner == uid {
		return errors.Trace(ErrInvalidOperator)
	}

	// 要退出的是管理员
	if g.existAdmin(gid, uid) == nil {
		if err = g.delAdmin(gid, uid); err != nil {
			return errors.Trace(err)
		}
	}

	return errors.Trace(g.delMember(gid, uid))
}

func (g *groupDB) invites(gid, operator int64, msg string, ids ...int64) error {
	//先获取下群信息，如果群不存在了就算了
	group, err := g.getGroupBase(gid)
	if err != nil {
		return errors.Trace(err)
	}

	//如果操作人不是管理员，也不是群主，就不能邀请人
	if err = g.existAdmin(gid, operator); err != nil && operator != group.Owner {
		return errors.Trace(ErrInvalidOperator)
	}

	for _, id := range ids {
		//已经进来了就算了
		if g.exist(gid, id) == nil {
			continue
		}

		if err := g.addInvite(gid, operator, id, msg); err != nil {
			return errors.Trace(err)
		}
	}
	return nil
}

// apply 申请入群
func (g *groupDB) apply(gid, uid int64, msg string) error {
	//先获取下群信息，如果群不存在了就算了
	_, err := g.getGroupBase(gid)
	if err != nil {
		return errors.Trace(err)
	}

	//已经进来了就算了
	if g.exist(gid, uid) == nil {
		return nil
	}

	if err := g.addApply(gid, uid, msg); err != nil {
		return errors.Trace(err)
	}
	return nil
}

//申请入群
type apply struct {
	Group int64
	User  int64
	Date  string
	Msg   string
}

//addApply 添加申请消息
func (g *groupDB) addApply(gid, uid int64, msg string) error {
	key := GroupApplyKey(gid, uid)
	log := apply{Group: gid, User: uid, Date: time.Now().Format(time.ANSIC), Msg: msg}

	data, err := json.Marshal(log)
	if err != nil {
		return errors.Trace(err)
	}

	return errors.Trace(g.db.Put(key, data, nil))
}

// existApply 这个用户是否申请过
func (g *groupDB) existApply(gid, uid int64) error {
	_, err := g.db.Get(GroupApplyKey(gid, uid), nil)
	return err
}

// delApply 删除这个用户的邀请
func (g *groupDB) delApply(gid, uid int64) error {
	return g.db.Delete(GroupApplyKey(gid, uid), nil)
}

type invite struct {
	Group    int64
	Operator int64
	User     int64
	Date     string
	Msg      string
}

//addInvite 添加邀请消息
func (g *groupDB) addInvite(gid, operator, uid int64, msg string) error {
	key := GroupInviteKey(gid, uid)
	log := invite{Group: gid, Operator: operator, User: uid, Date: time.Now().Format(time.ANSIC), Msg: msg}

	data, err := json.Marshal(log)
	if err != nil {
		return errors.Trace(err)
	}

	return errors.Trace(g.db.Put(key, data, nil))
}

// existInvite 是否邀请过这个用户
func (g *groupDB) existInvite(gid, uid int64) error {
	_, err := g.db.Get(GroupInviteKey(gid, uid), nil)
	return err
}

// delInvite 是否邀请过这个用户
func (g *groupDB) delInvite(gid, uid int64) error {
	return g.db.Delete(GroupInviteKey(gid, uid), nil)
}

// delUsers 批量踢人
func (g *groupDB) delUsers(gid, operator int64, ids ...int64) error {
	//先获取下群信息，如果群不存在了就算了
	group, err := g.getGroupBase(gid)
	if err != nil {
		return errors.Trace(err)
	}

	//如果操作人不是管理员，也不是群主，就不能删除其它人
	if err = g.existAdmin(gid, operator); err != nil && operator != group.Owner {
		return errors.Trace(ErrInvalidOperator)
	}

	for _, id := range ids {
		if err := g.delUser(group.Owner, gid, operator, id); err != nil {
			return errors.Trace(err)
		}
	}
	return nil
}

// delUser 踢人, 对象可能是管理员，也可能是普通群众
func (g *groupDB) delUser(owner, gid, operator, uid int64) error {
	//不能踢群主
	if owner == uid {
		return errors.Trace(ErrInvalidOperator)
	}

	// 要删除的对象是管理员
	if g.existAdmin(gid, uid) == nil {
		//除了群主其它人不能踢管理员
		if operator != owner {
			return errors.Trace(ErrInvalidOperator)
		}

		if err := g.delAdmin(gid, uid); err != nil {
			return errors.Trace(err)
		}
	}

	return errors.Trace(g.delMember(gid, uid))
}

// existAdmin 检测群中是否有这个管理员
func (g *groupDB) existAdmin(gid, uid int64) error {
	_, err := g.db.Get(GroupAdminKey(gid, uid), nil)
	return err
}

// exist 检测群中是否有这个用户
func (g *groupDB) exist(gid, uid int64) error {
	_, err := g.db.Get(GroupMemberKey(gid, uid), nil)
	return err
}

// agree 管理员同意入群
func (g *groupDB) agree(gid, operator, uid int64) error {
	//已经进来了, 因为多个管理员都会收到入群申请，会多次同意
	if g.exist(gid, uid) == nil {
		return nil
	}

	//先获取下群信息，如果群不存在了就算了
	group, err := g.getGroupBase(gid)
	if err != nil {
		return errors.Trace(err)
	}

	//如果操作人不是管理员，也不是群主
	if err = g.existAdmin(gid, operator); err != nil && operator != group.Owner {
		return errors.Trace(ErrInvalidOperator)
	}

	//没申请过不能通过
	if g.existApply(gid, uid) != nil {
		return errors.Trace(ErrInvalidOperator)
	}

	g.delApply(gid, uid)

	return g.addMember(gid, uid)
}

// accept 用户接受邀请入群
func (g *groupDB) accept(gid, uid int64) error {
	//先获取下群信息，如果群不存在了就算了
	_, err := g.getGroupBase(gid)
	if err != nil {
		return errors.Trace(err)
	}

	//已经进来了
	if g.exist(gid, uid) == nil {
		return nil
	}

	//如果没邀请过，不能接受邀请
	if g.existInvite(gid, uid) != nil {
		return errors.Trace(ErrInvalidOperator)
	}

	g.delInvite(gid, uid)

	return g.addMember(gid, uid)
}
