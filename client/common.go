package candy

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
