/*
	移动（坐标更新）路由业务
*/

package apis

import (
	"mmogameserver/core"
	"mmogameserver/pb"
	"tinyserver/tsinterface"
	"tinyserver/tsnet"

	"github.com/golang/protobuf/proto"
)

//Move 移动类
type Move struct {
	tsnet.BaseRouter
}

//Handler 业务处理方法
func (m *Move) Handler(request tsinterface.IRequest) {
	//解析客户端发送的proto协议消息（msgID:3）
	protoMsg := &pb.Position{}
	proto.Unmarshal(request.GetMsg().GetMsgData(), protoMsg)

	//通过连接属性获取当前玩家ID
	pid, _ := request.GetConnection().GetProperty("pid")

	//通过pid获取当前玩家对象
	player := core.WorldMgrObj.GetPlayerByPid(pid.(int32))

	//调用玩家对象的位置更新广播的方法（将当前的新坐标发送给周边全部玩家）
	player.UpdatePosition(protoMsg.X, protoMsg.Y, protoMsg.Z, protoMsg.V)
}
