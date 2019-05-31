/*
	世界聊天路由业务
*/

package apis

import (
	"fmt"
	"mmogameserver/core"
	"mmogameserver/pb"
	"tinyserver/tsinterface"
	"tinyserver/tsnet"

	"github.com/golang/protobuf/proto"
)

//WorldChat 世界聊天路由类
type WorldChat struct {
	tsnet.BaseRouter
}

//Handler 业务处理方法
func (wc *WorldChat) Handler(request tsinterface.IRequest) {
	//解析客户端发送额protibuf数据
	protoMsg := &pb.Talk{}
	if err := proto.Unmarshal(request.GetMsg().GetMsgData(), protoMsg); err != nil {
		fmt.Println("proto.Unmarshal error", err)
		return
	}
	//通过获取连接属性后获取当前玩家ID
	pid, err := request.GetConnection().GetProperty("pid")
	if err != nil {
		fmt.Println("GetProperty error", err)
		return
	}
	//通过玩家ID获取palyer对象
	player := core.WorldMgrObj.GetPlayerByPid(pid.(int32))

	//将当前玩家客户端发出的聊天数据广播给全部玩家
	player.SendTalkMsgToAll(protoMsg.GetContent())
}
