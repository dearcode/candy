/*
  
Candy是一款即时通信软件。  

Gate 接收客户端请求，负责客户端连接维护  
Notice 消息分发中心，整个系统的消息队列  
Store 消息及用户信息存储  
  
下载  
将 candy 代码 clone 到 $GOPATH/src/github.com/dearcode 目录下   
    
编译安装   
make  

运行  
依次启动 master, notice, store, gate, 直接运行不需要参数，默认使用9000到9004端口   


技术讨论QQ群：29996599


*/
package main

import (
	"fmt"

	"github.com/juju/errors"
	"github.com/syndtr/goleveldb/leveldb"

	"github.com/dearcode/candy/meta"
	"github.com/dearcode/candy/util"
	"github.com/dearcode/candy/util/log"
)

func main() {
	msg := meta.Message{}
	fmt.Printf("msg:%+v\n", msg)

	fmt.Printf("%s\n", util.EncodeInt64(123))
	log.Debug("xxx")

	err := fmt.Errorf("abc")
	err = errors.Trace(err)
	println(err)

	db := leveldb.DB{}
	fmt.Printf("%v\n", db)
}
