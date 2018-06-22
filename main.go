package main

import (
	"fmt"
	"github.com/antlinker/go-mqtt/client"
)

func main() {
	//开始日志　false则关闭日志显示
	client.Mlog.SetEnabled(true)
	client, err := client.CreateClient(client.MqttOption{
		Clientid:"test1232456789",
		Addr: "tcp://localhost:1883",
		UserName:"tab.zhang",
		Password:"password",
		KeepAlive:3600,
		//断开连接１秒后自动连接，０不自动重连
		ReconnTimeInterval: 1,
	})
	if err != nil {
		//配置文件解析失败
		panic("配置文件解析失败")
	}
	//建立连接
	err = client.Connect()
	if err != nil {
		//连接失败，不会进入自动重连状态
		panic(fmt.Errorf("连接失败:%v", err))
	}
	//断开连接
	client.Disconnect()
}
