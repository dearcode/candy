package candy

import (
	"encoding/json"
	"fmt"
)

//用户信息
type UserInfo struct {
	ID       int64
	Name     string
	NickName string
	Avatar   []byte
}

type UserList struct {
	Users []*UserInfo `json:"Users"`
}

//群组信息
type GroupInfo struct {
	ID    int64   `json:"ID"`
	Name  string  `json:"Name"`
	Users []int64 `json:"Users"`
}

type GroupList struct {
	Groups []*GroupInfo `json:"Groups"`
}

//好友信息
type FriendList struct {
	Users []int64 `json:"Users"`
}

func encodeJSON(data interface{}) ([]byte, error) {
	body, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	return body, nil
}

func DecodeUserInfo(data []byte) (*UserInfo, error) {
	userInfo := &UserInfo{}
	if err := json.Unmarshal(data, &userInfo); err != nil {
		return nil, fmt.Errorf("Decode UserInfo error:%v", err.Error())
	}
	return userInfo, nil
}

func DecodeUserList(data []byte) (*UserList, error) {
	userList := &UserList{}
	if err := json.Unmarshal(data, &userList); err != nil {
		return nil, fmt.Errorf("Decode UserList error:%v", err.Error())
	}
	return userList, nil
}

func DecodeFriendList(data []byte) (*FriendList, error) {
	friendList := &FriendList{}
	if err := json.Unmarshal(data, &friendList); err != nil {
		return nil, fmt.Errorf("Decode FriendList error:%v", err.Error())
	}
	return friendList, nil
}

func DecodeGroupList(data []byte) (*GroupList, error) {
	groupList := &GroupList{}
	if err := json.Unmarshal(data, &groupList); err != nil {
		return nil, fmt.Errorf("Decode GroupList error:%v", err.Error())
	}
	return groupList, nil
}
