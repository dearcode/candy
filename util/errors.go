package util

import (
	"github.com/juju/errors"
)

const (
	// ErrorOK 成功
	ErrorOK int32 = 0
	// ErrorFailure 未知错误
	ErrorFailure = 1
	// ErrorUserNameFormat 用户名格式错误
	ErrorUserNameFormat = 1000
	// ErrorUserNameLen 用户名长度错误
	ErrorUserNameLen = 1001
	// ErrorUserPasswdFormat 用户密码格式错误
	ErrorUserPasswdFormat = 1010
	// ErrorUserPasswdLen 用户密码长度错误
	ErrorUserPasswdLen = 1011
	// ErrorUserNickFormat 昵称格式错误
	ErrorUserNickFormat = 1020
	// ErrorUserNickLen 昵称长度错误
	ErrorUserNickLen = 1021
	// ErrorGetSession 获取回话失败
	ErrorGetSession = 1030
	// ErrorMasterNewID 生成ID失败
	ErrorMasterNewID = 1031
	// ErrorRegister 注册失败
	ErrorRegister = 1032
	// ErrorOffline 已经离线
	ErrorOffline = 1033
	// ErrorUpdateUserInfo 更新用户信息失败
	ErrorUpdateUserInfo = 1034
	// ErrorUpdateUserPasswd 更新用户密码失败
	ErrorUpdateUserPasswd = 1035
	// ErrorGetUserInfo 获取用户信息失败
	ErrorGetUserInfo = 1036
	// ErrorAuth 用户认证失败（提示用户名或密码错误）
	ErrorAuth = 1037
	// ErrorSubscribe 消息订阅失败
	ErrorSubscribe = 1038
	// ErrorUnSubscribe 取消消息订阅失败
	ErrorUnSubscribe = 1039
	// ErrorNewMessage 发送消息失败
	ErrorNewMessage = 1040
	// ErrorFriendSelf 不能添加自己为好友
	ErrorFriendSelf = 1041
	// ErrorAddFriend 添加好友失败
	ErrorAddFriend = 1042
	// ErrorGetOnlineSession 用户不在线
	ErrorGetOnlineSession = 1043
	// ErrorLoadFriendList 加载好友列表失败
	ErrorLoadFriendList = 1044
	// ErrorFindUser 查找用户失败
	ErrorFindUser = 1045
	// ErrorCreateGroup 创建分组失败
	ErrorCreateGroup = 1046
	// ErrorLoadGroup 加载群组列表失败
	ErrorLoadGroup = 1047
	// ErrorUploadFile 上传文件失败
	ErrorUploadFile = 1048
	// ErrorCheckFile 文件检查失败
	ErrorCheckFile = 1049
	// ErrorDownloadFile 下载文件失败
	ErrorDownloadFile = 1050
	// ErrorLoadMessage 加载消息列表失败
	ErrorLoadMessage = 1051
	// ErrorUpdateSignature 更新签名失败
	ErrorUpdateSignature = 1052
)

var (
	// ErrInvalidContext 从context中解析客户端地址时出错.
	ErrInvalidContext = errors.New("invalid context")
)
