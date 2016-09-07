![Logo](/res/logo.png?raw=true "Candy logo")

[![Circle CI](https://circleci.com/gh/dearcode/candy.svg?style=svg)](https://circleci.com/gh/dearcode/candy) 


## 项目背景 
  Candy是邮差网开源的一款即时通信软件，最初是为自己内部沟通而做的 IM 工具。 

## 项目框架
  服务端所在server目录中 
  Gate: 接收客户端请求，负责客户端连接维护 
  Notice: 消息分发中心，整个系统的消息队列 
  Store: 消息及用户信息存储 

## 编译安装 
  `make` 
  依次启动 `master`, `notice`, `store`, `gate`， 直接运行不需要参数，默认使用9000到9004端口 


技术讨论QQ群：[![Circle CI](http://pub.idqqimg.com/wpa/images/group.png)](http://shang.qq.com/wpa/qunwpa?idkey=d43cad7db88d71f70da81523c02b2fe59343111e1d0a9d5f5ac2a198ee047279)


