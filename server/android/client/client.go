package main

import (
	"bufio"
	"fmt"
	"os"

	candy "github.com/dearcode/candy/server/android"
	"github.com/dearcode/candy/server/util/log"
)

func main() {
	c := candy.NewCandyClient("127.0.0.1:9000")
	if err := c.Start(); err != nil {
		log.Errorf("start client error:%s", err.Error())
		return
	}

	fmt.Println("---------------------------------")
	fmt.Println("1. 注册用户")
	fmt.Println("2. 登陆")
	fmt.Println("3. 更新用户信息")
	fmt.Println("4. 查找用户")
	fmt.Println("5. 添加好友")
	fmt.Println("0. 退出")
	fmt.Println("---------------------------------")

	running := true
	reader := bufio.NewReader(os.Stdin)
	for running {
		data, _, _ := reader.ReadLine()
		command := string(data)
		log.Debugf("command", command)
		if command == "0" {
			running = false
		} else if command == "1" {
			fmt.Println("请输入用户名:")
			data, _, _ = reader.ReadLine()
			userName := string(data)
			fmt.Println("请输入密码:")
			data, _, _ = reader.ReadLine()
			userPassword := string(data)

			id, err := c.Register(userName, userPassword)
			if err != nil {
				log.Errorf("Register error:%v", err)
				continue
			}

			log.Debugf("Register success, userID:%v", id)
		}
	}

}
