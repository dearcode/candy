![Logo](https://raw.githubusercontent.com/dearcode/web/master/static/img/logo.png "Candy logo")

[![Build Status](https://travis-ci.org/dearcode/candy.svg?branch=master)](https://travis-ci.org/dearcode/candy)
[![Go Report Card](https://goreportcard.com/badge/github.com/dearcode/candy)](https://goreportcard.com/report/github.com/dearcode/candy)
[![GoDoc](https://godoc.org/github.com/dearcode/candy?status.svg)](https://godoc.org/github.com/dearcode/candy)


## 项目背景 
  Candy是新一代即时通信软件, 我们是一套完整的即时通讯开源解决方案，包括移动端，WEB端，PC端。  

## Candy架构
![candy server](https://raw.githubusercontent.com/dearcode/doc/master/candy_server.jpg "Candy Server")  
  * Gate 接收客户端请求，负责客户端连接维护  
  * Master 负责全局ID生成，及节点管理等功能  
  * Notifer 消息分发中心，整个系统的消息队列, 负责在线及离线消息推送   
  * Store 逻辑处理中心，单机版中兼顾存储服务  

### Android  
    https://github.com/dearcode/candy-android  

### Web  
    https://github.com/dearcode/web  

### CUI  
    https://github.com/zeazen/candy-cui  
    
## 项目运行
### 获取源码
  `mkdir -p $GOPATH/src/github.com/dearcode`  
  `cd $GOPATH/src/github.com/dearcode`  
  `git clone git@github.com:dearcode/candy.git`  
   
### 编译安装 
  `make`  
  依次启动 `master`, `notice`, `store`, `gate`  
  直接运行不需要参数，默认使用9000到9004端口   

###  测试运行
  项目根目录中的candy.go是一个命令行的全功能展示  
  `go run candy.go`

## 跨平台支持  
  Candy客户端可以直接使用gomobile编译出安卓/IOS客户端可用的库文件，可以方便快捷的和服务器通信, 用户无需关注客户端服务器的实现； 只需要把重点放在客户端开发上。  

### Android库编译  
  可以直接使用gomobile编译出安卓可用的版本  

#### 环境要求：  
  1. JDK 1.7版本以上  
  2. Android SDK  
  3. 安装gomobile   

#### 编译命令:    
  `gomobile bind -v -target=android/arm github.com/dearcode/candy/client`  
  
### IOS库编译
  可以直接使用gomobile编译出IOS可用的版本  

#### 环境要求：  
  1. IOS SDK  
  2. 安装gomobile   

#### 编译命令:   
  `gomobile bind -v -target=ios github.com/dearcode/candy/client`   



技术讨论QQ群：[![Circle CI](http://pub.idqqimg.com/wpa/images/group.png)](http://shang.qq.com/wpa/qunwpa?idkey=d43cad7db88d71f70da81523c02b2fe59343111e1d0a9d5f5ac2a198ee047279) 29996599    


