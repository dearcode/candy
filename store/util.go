package store

import (
	"math"

	"github.com/dearcode/candy/server/util"
)

// UserMessageKey create msg key by userid and msgid.
func UserMessageKey(user int64, msg int64) []byte {
	return util.EncodeKey(user, util.UserMessagePrefix, msg)
}

// UserLastMessageKey create max user message key.
func UserLastMessageKey(id int64) []byte {
	return util.EncodeKey(id, util.UserLastMessagePrefix)
}

// UserGroupRange create user group key range.
func UserGroupRange(id int64) ([]byte, []byte) {
	return util.EncodeKey(id, util.UserGroupPrefix, int64(0)), util.EncodeKey(id, util.UserGroupPrefix, math.MaxInt64)
}

// UserFriendKey create user firend key.
func UserFriendKey(user int64, friend int64) []byte {
	return util.EncodeKey(user, util.UserFriendPrefix, friend)
}

// UserFriendRange create user friend key range.
func UserFriendRange(id int64) ([]byte, []byte) {
	return util.EncodeKey(id, util.UserFriendPrefix, int64(0)), util.EncodeKey(id, util.UserFriendPrefix, math.MaxInt64)
}
