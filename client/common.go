package candy

//用户信息
type UserInfo struct {
	ID       int64
	Name     string
	NickName string
	Avatar   []byte
}

type UserList struct {
	Users     []*UserInfo
	JsonUsers string //由于当前不支持slice，所以先使用Json字符串代替
}

//群组信息
type GroupInfo struct {
	ID        int64
	Name      string
	Users     []int64
	JsonUsers string //由于当前不支持slice，所以先使用Json字符串代替
}

type GroupList struct {
	Groups     []*GroupInfo
	JsonGroups string //由于当前不支持slice，所以先使用Json字符串代替
}

//好友信息
type FriendList struct {
	Users     []int64
	JsonUsers string //由于当前不支持slice，所以先使用Json字符串代替
}
