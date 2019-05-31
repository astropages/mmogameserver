package main

import (
	"fmt"
	"mmogameserver/apis"
	"mmogameserver/core"
	"tinyserver/tsinterface"
	"tinyserver/tsnet"
)

//OnConnectionAdd 客户端建立建立之后触发的hook函数
func OnConnectionAdd(conn tsinterface.IConnection) {
	fmt.Println("玩家连接...")

	//创建一个玩家（将连接和玩家模块绑定）
	p := core.NewPlayer(conn)

	//给客户端发送msgID:1（玩家ID）
	p.ReturnPid()

	//给客户端发送msgID:200（玩家坐标）
	p.ReturnPlayerPosition()

	//上线成功，将玩家对象添加到世界管理器中
	core.WorldMgrObj.AddPlayer(p)

	//给连接添加一个键为pid的属性
	conn.SetProperty("pid", p.Pid)

	//同步周边玩家，广播当前玩家的位置信息
	p.SyncSurrounding()

	fmt.Printf("玩家%d上线，当前共%d个玩家\n", p.Pid, len(core.WorldMgrObj.Players))

}

//OnConnectionLost 客户端连接断开前触发的hook函数
func OnConnectionLost(conn tsinterface.IConnection) {
	//获取当前下线的玩家
	pid, _ := conn.GetProperty("pid")
	player := core.WorldMgrObj.GetPlayerByPid(pid.(int32))
	//调用玩家下线广播的方法
	player.OffLine()
}

func main() {
	//创建一个服务器
	s := tsnet.NewServer("MMO Game Server")

	//注册连接建立后的hook函数
	s.AddOnConnStart(OnConnectionAdd)
	//注册连接断开前的hook函数
	s.AddOnConnStop(OnConnectionLost)

	//定义msgID为2的业务路由（世界聊天）
	s.AddRouter(2, &apis.WorldChat{})
	//定义msgID为3的业务路由（坐标更新）
	s.AddRouter(3, &apis.Move{})

	//启动服务
	s.Serve()
}
