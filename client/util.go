package candy

import (
	"encoding/json"
	"fmt"
	"regexp"
)

var (
	minUsernameLen   = 6
	minUserpasswdLen = 6
)

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
func CheckUserName(name string) error {
	if len(name) < minUsernameLen {
		return fmt.Errorf("UserName minimum length is %v", minUsernameLen)
	}

	reg := regexp.MustCompile(`^[a-zA-Z0-9_-]+@[a-zA-Z0-9_-]+(\.[a-zA-Z0-9_-]+)+$`)
	if !reg.MatchString(name) {
		return fmt.Errorf("UserName format error, just support email address")
	}

	return nil
}

// CheckUserPassword - 用户密码校验， 密码强度暂时不限制， 当前只限制密码最小长度
func CheckUserPassword(passwd string) error {
	if len(passwd) < minUserpasswdLen {
		return fmt.Errorf("UserPasswd minimum length is %v", minUserpasswdLen)
	}

	//TODO 密码复杂度校验，当前为了方便测试先不加

	return nil
}

// CheckNickName - 用户昵称校验
func CheckNickName(nick string) error {
	//TODO 后续完善
	return nil
}
