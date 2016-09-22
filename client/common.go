package candy

//用户信息
type UserInfo struct {
	ID       int64
	Name     string
	NickName string
	Avatar   []byte
}

type UserList struct {
	Users []*UserInfo
}

//群组信息
type GroupInfo struct {
	ID    int64
	Name  string
	Users []int64
}

type GroupList struct {
	Groups []*GroupInfo
}
