package store

import (
	"math"

	"github.com/dearcode/candy/meta"
	"github.com/dearcode/candy/util"
)

func UserMessageKey(user int64, msg int64) []byte {
	return util.EncodeKey(user, util.UserMessagePrefix, msg)
}

func UserLastMessageKey(id int64) []byte {
	return util.EncodeKey(id, util.UserLastMessagePrefix)
}

func UserGroupRange(id int64) ([]byte, []byte) {
	return util.EncodeKey(id, util.UserGroupPrefix, int64(0)), util.EncodeKey(id, util.UserGroupPrefix, math.MaxInt64)
}

func UserFriendKey(user int64, friend int64) []byte {
	return util.EncodeKey(user, util.UserFriendPrefix, friend)
}

func UserFriendRange(id int64) ([]byte, []byte) {
	return util.EncodeKey(id, util.UserFriendPrefix, int64(0)), util.EncodeKey(id, util.UserFriendPrefix, math.MaxInt64)
}


