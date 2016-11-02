package util

import (
	"crypto/md5"
	"fmt"
	"strings"
	"time"

	"github.com/juju/errors"
	"golang.org/x/net/context"
	"google.golang.org/grpc/metadata"
)

var (
	// BuildTime 编译时间.
	BuildTime = ""
	// BuildVersion git版本.
	BuildVersion = ""
)

const (
	// UserDBPath 用户数据库位置.
	UserDBPath = "user"
	// MessageDBPath 消息数据库位置.
	MessageDBPath = "message"
	// GroupDBPath 群组数据库位置.
	GroupDBPath = "group"

	// FileBlockPath 块文件存储位置
	FileBlockPath = "file_block"
	// FileDBPath 文件索引存储位置
	FileDBPath = "file"

	// MessageLogDBPath 消息记录数据库位置.
	MessageLogDBPath = "message_log"

	// UserMessagePrefix 用户消息前缀.
	UserMessagePrefix = int64(0)

	//GroupAdminPrefix 群管理员
	GroupAdminPrefix = int64(0)
	//GroupMemberPrefix 群成员
	GroupMemberPrefix = int64(1)
	//GroupInvitePrefix 群邀请消息
	GroupInvitePrefix = int64(2)
	//GroupApplyPrefix 申请入群消息
	GroupApplyPrefix = int64(3)

	// UserLastMessagePrefix 用户最后一个消息ID.
	UserLastMessagePrefix = int64(1)
	// UserGroupPrefix 用户组前缀.
	UserGroupPrefix = int64(2)
	// UserFriendPrefix 用户好友前缀.
	UserFriendPrefix = int64(3)
	// UserIDPrefix 用户前缀.
	UserIDPrefix = int64(4)

	// NetworkTimeout 超时.
	NetworkTimeout = time.Second * 3
)

// PrintVersion 输出当前程序编译信息.
func PrintVersion() {
	fmt.Printf("Candy\n")
	fmt.Printf("Build Time: %s\n", BuildTime)
	fmt.Printf("Git Version: %s\n", BuildVersion)
}

// EncodeInt64 编码int64到[]byte.
func EncodeInt64(v int64) []byte {
	return []byte{byte(v >> 56), byte(v >> 48), byte(v >> 40), byte(v >> 32), byte(v >> 24), byte(v >> 16), byte(v >> 8), byte(v)}
}

// DecodeInt64 解码[]byte到int64.
func DecodeInt64(b []byte) int64 {
	if len(b) < 8 {
		return 0
	}

	return int64(b[0])<<56 + int64(b[1])<<48 + int64(b[2])<<40 + int64(b[3])<<32 + int64(b[4])<<24 + int64(b[5])<<16 + int64(b[6])<<8 + int64(b[7])
}

// EncodeKey 编码int64数组到[]byte.
func EncodeKey(args ...int64) []byte {
	var key []byte
	for _, v := range args {
		key = append(key, EncodeInt64(v)...)
	}
	return key
}

// MD5 计算md5.
func MD5(data []byte) []byte {
	hash := md5.New()
	hash.Write(data)
	return hash.Sum(nil)
}

// ContextGet get value from context's metadata.
func ContextGet(ctx context.Context, key string) (string, error) {
	md, ok := metadata.FromContext(ctx)
	if !ok {
		return "", errors.Trace(ErrInvalidContext)
	}

	vals, ok := md[key]
	if !ok {
		return "", errors.Trace(ErrInvalidContext)
	}

	if len(vals) != 1 {
		return "", errors.Trace(ErrInvalidContext)
	}

	return vals[0], nil
}

// ContextSet set value to context's metadata.
func ContextSet(ctx context.Context, key, val string) context.Context {
	md, ok := metadata.FromContext(ctx)
	if !ok {
		return metadata.NewContext(ctx, metadata.Pairs(key, val))
	}
	md[key] = []string{val}
	return ctx
}

//Split 如果s为空字符串返回空数组
func Split(s, sep string) []string {
	if len(s) == 0 {
		return nil
	}
	return strings.Split(s, sep)
}
