/*
	玩家模块
*/

package core

import (
	"fmt"
	"math/rand"
	"mmogameserver/pb"
	"sync"
	"tinyserver/tsinterface"

	"github.com/golang/protobuf/proto"
)

//Player 玩家类
type Player struct {
	Pid  int32                   //玩家ID
	Conn tsinterface.IConnection //当前玩家的连接
	X    float32                 //平面X轴坐标
	Y    float32                 //高度
	Z    float32                 //平面Y轴坐标
	V    float32                 //玩家的朝向
}

//PidGen playerID生成器
var PidGen int32 = 1 //玩家ID计数

//IDLock 保护PidGen生成器的互斥锁
var IDLock sync.Mutex

//NewPlayer 初始化玩家
func NewPlayer(conn tsinterface.IConnection) *Player {
	//分配一个玩家ID
	IDLock.Lock()
	id := PidGen
	PidGen++
	IDLock.Unlock()

	//创建一个玩家对象
	p := &Player{
		Pid:  id,
		Conn: conn,
		X:    float32(160 + rand.Intn(10)), //随机生成玩家上线所在的X轴坐标
		Y:    0,
		Z:    float32(140 + rand.Intn(10)), //随机在140坐标附近Y轴坐标上线
		V:    0,                            //角度为0
	}

	return p
}

//SendMsg 玩家和对端客户端发送消息的方法
func (p *Player) SendMsg(msgID uint32, protoStruct proto.Message) error {
	//将proto结构体数据转换为二进制数据
	banaryProtoData, err := proto.Marshal(protoStruct)
	if err != nil {
		fmt.Println("proto.Marshal error ", err)
		return err
	}

	//调用服务器框架里连接模块给客户端发送数据的方法
	err = p.Conn.Send(msgID, banaryProtoData)
	if err != nil {
		fmt.Println("p.Conn.Send error!", err)
		return err
	}

	return nil
}

//ReturnPid 服务器给客户端发送一个玩家初始ID的方法
func (p *Player) ReturnPid() {
	//定义proto数据：协议里的玩家ID结构体
	protoMsg := &pb.SyncPid{
		Pid: p.Pid,
	}
	//将proto数据作为msgID为1的消息发送给客户端
	p.SendMsg(1, protoMsg)
}

//ReturnPlayerPosition 服务器给客户端发送一个玩家的初始化位置信息的方法
func (p *Player) ReturnPlayerPosition() {
	//构建msgID为200的消息
	protoMsg := &pb.BroadCast{
		Pid: p.Pid,
		Tp:  2, //坐标信息
		Data: &pb.BroadCast_P{
			P: &pb.Position{
				X: p.X,
				Y: p.Y,
				Z: p.Z,
				V: p.V,
			},
		},
	}

	//将proto数据作为msgID为200的消息发送给客户端
	p.SendMsg(200, protoMsg)
}

//SendTalkMsgToAll 将聊天数据广播给全部在线玩家的方法
func (p *Player) SendTalkMsgToAll(content string) {
	//定义一个符合广播协议数据的protoMsg
	protoMsg := &pb.BroadCast{
		Pid: p.Pid,
		Tp:  1,
		Data: &pb.BroadCast_Content{
			Content: content,
		},
	}

	//获取全部在线玩家
	players := WorldMgrObj.GetAllPlayers()

	//将protoMsg作为msgID为200的数据广播给全部玩家
	for _, player := range players {
		player.SendMsg(200, protoMsg)
	}
}

//GetSurroundingPlayers 获取当前玩家周边的全部玩家
func (p *Player) GetSurroundingPlayers() []*Player {
	pids := WorldMgrObj.AoiMgr.GetSurroundPIDsByPos(p.X, p.Z)
	players := make([]*Player, 0, len(pids))
	for _, pid := range pids {
		players = append(players, WorldMgrObj.GetPlayerByPid(int32(pid)))
	}
	return players
}

//SyncSurrounding 将自己的消息同步给周围玩家
func (p *Player) SyncSurrounding() {
	//获取当前玩家周边九宫格范围内的全部玩家
	players := p.GetSurroundingPlayers()

	//定义一个当前玩家ID和位置的protobuf消息
	protoMsg := &pb.BroadCast{
		Pid: p.Pid,
		Tp:  2,
		Data: &pb.BroadCast_P{
			P: &pb.Position{
				X: p.X,
				Y: p.Y,
				Z: p.Z,
				V: p.V,
			},
		},
	}

	//将玩家ID和位置的protobuf消息作为msgID为200的数据发送给周边玩家
	for _, player := range players {
		player.SendMsg(200, protoMsg)
	}

	//将周围其他玩家的ID和位置信息发送给当前玩家

	//当前玩家周边全部玩家的信息集合
	playersProtoMsg := make([]*pb.Player, 0, len(players))
	for _, player := range players {
		//定义一个player的protobuf协议消息
		p := &pb.Player{
			Pid: player.Pid,
			P: &pb.Position{
				X: player.X,
				Y: player.Y,
				Z: player.Z,
				V: player.V,
			},
		}
		playersProtoMsg = append(playersProtoMsg, p)
	}

	//定义一个周边全部玩家的protobuf协议消息
	syncPlayersProtoMsg := &pb.SyncPlayers{
		Ps: playersProtoMsg[:],
	}

	//将周边全部玩家的protobuf消息作为msgID为202的数据发送给当前客户端
	p.SendMsg(202, syncPlayersProtoMsg)
}

//OnExchangeAoiGrid 格子切换时视野处理的方法
func (p *Player) OnExchangeAoiGrid(oldGrid, newGrid int) {
	//获取九宫格成员
	oldGrids := WorldMgrObj.AoiMgr.GetSurroundGridsByGid(oldGrid)
	//建立旧九宫格成员哈希表用于快速查找
	oldGridsMap := make(map[int]bool, len(oldGrids))
	for _, grid := range oldGrids {
		oldGridsMap[grid.GID] = true
	}
	//获取新九宫格成员
	newGrids := WorldMgrObj.AoiMgr.GetSurroundGridsByGid(newGrid)
	//建立新九宫格成员哈希表用于快速查找
	newGridsMap := make(map[int]bool, len(newGrids))
	for _, grid := range newGrids {
		newGridsMap[grid.GID] = true
	}

	//---视野消失处理---

	//构建msgID:201
	offlineMsg := &pb.SyncPid{
		Pid: p.Pid,
	}

	//获取旧九宫格中存在，新九宫格中不存在的格子
	leavingGrids := make([]*Grid, 0)
	for _, grid := range oldGrids {
		if _, ok := newGridsMap[grid.GID]; !ok {
			leavingGrids = append(leavingGrids, grid)
		}
	}

	//获取leavingGrids中的全部玩家
	for _, grid := range leavingGrids {
		players := WorldMgrObj.GetPlayersByGrid(grid.GID)

		for _, player := range players {
			//将自己从其它玩家的客户端中消失
			player.SendMsg(201, offlineMsg)
			//将其它玩家从自己的客户端中消失
			anotherOfflineMsg := &pb.SyncPid{
				Pid: player.Pid,
			}
			p.SendMsg(201, anotherOfflineMsg)
		}
	}

	//---视野出现处理---

	onlineMsg := &pb.BroadCast{
		Pid: p.Pid,
		Tp:  2,
		Data: &pb.BroadCast_P{
			P: &pb.Position{
				X: p.X,
				Y: p.Y,
				Z: p.Z,
				V: p.V,
			},
		},
	}

	//获取新九宫格中存在，旧九宫格中不存在的格子
	enteringGrids := make([]*Grid, 0)
	for _, grid := range newGrids {
		if _, ok := oldGridsMap[grid.GID]; !ok {
			enteringGrids = append(enteringGrids, grid)
		}
	}

	//获取需要显示视野的格子集合中的全部玩家，然后分别进行消息发送
	for _, grid := range enteringGrids {
		players := WorldMgrObj.GetPlayersByGrid(grid.GID)

		for _, player := range players {
			//让自己出现在其它玩家的视野中
			player.SendMsg(200, onlineMsg)
			//让其它玩家出现在自己的视野中
			anoterOnlineMsg := &pb.BroadCast{
				Pid: player.Pid,
				Tp:  2,
				Data: &pb.BroadCast_P{
					P: &pb.Position{
						X: player.X,
						Y: player.Y,
						Z: player.Z,
						V: player.V,
					},
				},
			}
			p.SendMsg(200, anoterOnlineMsg)
		}
	}

}

//UpdatePosition 位置更新广播（将当前的新坐标发送给周边全部玩家）的方法
func (p *Player) UpdatePosition(x, y, z, v float32) {
	//判断当前玩家是否跨越格子
	//旧格子ID
	oldGrid := WorldMgrObj.AoiMgr.GetGridByPos(p.X, p.Z)
	//新格子ID
	NewGrid := WorldMgrObj.AoiMgr.GetGridByPos(x, z)
	//触发格子切换
	if oldGrid != NewGrid {
		//把pid从旧的aoi格子中删除
		WorldMgrObj.AoiMgr.RemovePidFromGrid(int(p.Pid), oldGrid)
		//把pid添加到新的aoi格子
		WorldMgrObj.AoiMgr.AddPidToGrid(int(p.Pid), NewGrid)
		//视野处理
		p.OnExchangeAoiGrid(oldGrid, NewGrid)
	}

	//将新坐标更新到当前玩家
	p.X = x
	p.Y = y
	p.Z = z
	p.V = v

	//定义玩家ID和位置的proto协议消息
	protoMsg := &pb.BroadCast{
		Pid: p.Pid,
		Tp:  4, //更新坐标
		Data: &pb.BroadCast_P{
			P: &pb.Position{
				X: p.X,
				Y: p.Y,
				Z: p.Z,
				V: p.V,
			},
		},
	}

	//获取当前玩家周边九宫格范围内的全部玩家
	players := p.GetSurroundingPlayers()

	//将玩家ID和位置的proto协议消息作为msgID为200的数据发送给每个玩家的客户端
	for _, player := range players {
		player.SendMsg(200, protoMsg)
	}
}

//OffLine 玩家下线广播的方法
func (p *Player) OffLine() {
	//获取当前玩家周边的全部玩家
	players := p.GetSurroundingPlayers()

	//定义一个玩家ID的proto协议消息
	protoMsg := &pb.SyncPid{
		Pid: p.Pid,
	}

	//将玩家ID的proto协议消息以msgID为201的数据发送给周边全部玩家
	for _, player := range players {
		player.SendMsg(201, protoMsg)
	}

	//将下线的玩家从世界管理器中移除
	WorldMgrObj.RemovePlayerByPid(p.Pid)

	//将下线玩家从地图AOIManager中移除
	WorldMgrObj.AoiMgr.RemoveFromGridByPos(int(p.Pid), p.X, p.Z)

}
