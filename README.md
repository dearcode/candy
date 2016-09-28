![Logo](https://raw.githubusercontent.com/dearcode/web/master/logo.png "Candy logo")

[![Build Status](https://travis-ci.org/dearcode/candy.svg?branch=master)](https://travis-ci.org/dearcode/candy)
[![Go Report Card](https://goreportcard.com/badge/github.com/dearcode/candy)](https://goreportcard.com/report/github.com/dearcode/candy)
[![GoDoc](https://godoc.org/github.com/dearcode/candy?status.svg)](https://godoc.org/github.com/dearcode/candy)


## 项目背景 
  Candy是一款即时通信软件。  

## 项目框架 
### 服务端  
  * Gate 接收客户端请求，负责客户端连接维护  
  * Notice 消息分发中心，整个系统的消息队列  
  * Store 消息及用户信息存储  

### Android客户端
    https://github.com/dearcode/candy-android  
    
## 项目运行
### 获取源码
  推荐使用以下方式之一获得源码:
  1. 将 candy 代码 clone 到 $GOPATH/src/github.com/dearcode 目录下 
  2. 使用 go get -u github.com/dearcode/candy
  以保证 import 路径正确  
   
### 编译安装 
  `make`  
  依次启动 `master`, `notice`, `store`, `gate`  
  直接运行不需要参数，默认使用9000到9004端口   


技术讨论QQ群：[![Circle CI](http://pub.idqqimg.com/wpa/images/group.png)](http://shang.qq.com/wpa/qunwpa?idkey=d43cad7db88d71f70da81523c02b2fe59343111e1d0a9d5f5ac2a198ee047279)


