package util

import (
	"fmt"
	"time"
)

var (
	BUILD_TIME    = ""
	BUILD_VERSION = ""
)

const (
	UserDBPath       = "user"
	MessageDBPath    = "message"
	GroupDBPath      = "group"
	MessageLogDBPath = "message_log"

	UserMessagePrefix     = int64(0)
	UserLastMessagePrefix = int64(1)
	UserGroupPrefix       = int64(2)
	UserFriendPrefix      = int64(3)

	NetworkTimeout = time.Second * 3
)

func PrintVersion() {
	fmt.Printf("Candy\n")
	fmt.Printf("Build Time: %s\n", BUILD_TIME)
	fmt.Printf("Git Version: %s\n", BUILD_VERSION)
}

func EncodeInt64(v int64) []byte {
	return []byte{byte(v >> 56), byte(v >> 48), byte(v >> 40), byte(v >> 32), byte(v >> 24), byte(v >> 16), byte(v >> 8), byte(v)}
}

func DecodeInt64(b []byte) int64 {
	if len(b) < 8 {
		return 0
	}

	return int64(b[0])<<56 + int64(b[1])<<48 + int64(b[2])<<40 + int64(b[3])<<32 + int64(b[4])<<24 + int64(b[5])<<16 + int64(b[6])<<8 + int64(b[7])
}

func EncodeKey(args ...int64) []byte {
	var key []byte
	for _, v := range args {
		key = append(key, EncodeInt64(v)...)
	}
	return key
}
