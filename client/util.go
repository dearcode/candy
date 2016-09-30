package client

import (
	"encoding/json"
	"fmt"
	"regexp"

	"github.com/dearcode/candy/util"
)

var (
	minUsernameLen   = 6
	minUserpasswdLen = 6
)

// Error 返回错误
type Error struct {
	Code int32
	Msg  string
}

// NewError - create an new Error
func NewError(code int32, msg string) *Error {
	return &Error{Code: code, Msg: msg}
}

// ErrorParse - parse error string to an Error object
func ErrorParse(msg string) *Error {
	var e Error
	if err := json.Unmarshal([]byte(msg), &e); err != nil {
		e.Msg = msg
		return &e
	}
	return &e
}

// Error - implement error interface
func (e *Error) Error() string {
	data, err := json.Marshal(e)
	if err != nil {
		return err.Error()
	}

	return string(data)
}

func encodeJSON(data interface{}) ([]byte, error) {
	body, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	return body, nil
}

// DecodeUserInfo 把json数据解析成UserInfo
func DecodeUserInfo(data []byte) (*UserInfo, error) {
	userInfo := &UserInfo{}
	if err := json.Unmarshal(data, &userInfo); err != nil {
		return nil, fmt.Errorf("Decode UserInfo error:%v", err.Error())
	}
	return userInfo, nil
}

// DecodeUserList 把json数据解析成UserList
func DecodeUserList(data []byte) (*UserList, error) {
	userList := &UserList{}
	if err := json.Unmarshal(data, &userList); err != nil {
		return nil, fmt.Errorf("Decode UserList error:%v", err.Error())
	}
	return userList, nil
}

// DecodeFriendList 把json数据解析成FriendList
func DecodeFriendList(data []byte) (*FriendList, error) {
	friendList := &FriendList{}
	if err := json.Unmarshal(data, &friendList); err != nil {
		return nil, fmt.Errorf("Decode FriendList error:%v", err.Error())
	}
	return friendList, nil
}

// DecodeGroupList 把json数据解析成GroupList
func DecodeGroupList(data []byte) (*GroupList, error) {
	groupList := &GroupList{}
	if err := json.Unmarshal(data, &groupList); err != nil {
		return nil, fmt.Errorf("Decode GroupList error:%v", err.Error())
	}
	return groupList, nil
}

// CheckUserName - 用户名校验， 用户名目前只支持邮箱, 长度至少6位
func CheckUserName(name string) (int32, error) {
	if len(name) < minUsernameLen {
		return util.ErrorUserNameLen, fmt.Errorf("UserName minimum length is %v", minUsernameLen)
	}

	reg := regexp.MustCompile(`^[a-zA-Z0-9_-]+@[a-zA-Z0-9_-]+(\.[a-zA-Z0-9_-]+)+$`)
	if !reg.MatchString(name) {
		return util.ErrorUserNameFormat, fmt.Errorf("UserName format error, just support email address")
	}

	return util.ErrorOK, nil
}

// CheckUserPassword - 用户密码校验， 密码强度暂时不限制， 当前只限制密码最小长度
func CheckUserPassword(passwd string) (int32, error) {
	if len(passwd) < minUserpasswdLen {
		return util.ErrorUserPasswdLen, fmt.Errorf("UserPasswd minimum length is %v", minUserpasswdLen)
	}

	//TODO 密码复杂度校验，当前为了方便测试先不加
	return util.ErrorOK, nil
}

// CheckNickName - 用户昵称校验
func CheckNickName(nick string) (int32, error) {
	//TODO 后续完善
	return util.ErrorOK, nil
}
