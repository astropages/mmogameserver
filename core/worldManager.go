/*
	当前场景的世界管理模块
*/

package core

import "sync"

//世界地图的边界参数
const (
	AoiMinX    int = 85
	AoiMaxX    int = 410
	AoiCountsX int = 10
	AoiMinY    int = 75
	AoiMaxY    int = 400
	AoiCountsY int = 20
)

//WorldManager 世界管理模块类
type WorldManager struct {
	Players map[int32]*Player //当前全部在线的Player集合
	PLock   sync.RWMutex      //保护Player集合的锁
	AoiMgr  *AOIManager       //当前地图的AOI管理器
}

//WorldMgrObj 对外提供一个全局世界管理模块指针
var WorldMgrObj *WorldManager

func init() {
	//创建一个全局的世界管理模块对象
	WorldMgrObj = NewWorldManager()
}

//NewWorldManager 初始化世界管理模块对象
func NewWorldManager() *WorldManager {
	wm := &WorldManager{
		AoiMgr:  NewAOIManager(AoiMinX, AoiMaxX, AoiCountsX, AoiMinY, AoiMaxY, AoiCountsY),
		Players: make(map[int32]*Player),
	}
	return wm
}

//AddPlayer 添加一个玩家的方法
func (wm *WorldManager) AddPlayer(player *Player) {
	//加入世界管理器中
	wm.PLock.Lock()
	wm.Players[player.Pid] = player
	wm.PLock.Unlock()

	//加入到世界地图中
	wm.AoiMgr.AddToGridByPos(int(player.Pid), player.X, player.Z)
}

//RemovePlayerByPid 删除一个玩家的方法
func (wm *WorldManager) RemovePlayerByPid(pid int32) {
	//从世界管理器中删除
	wm.PLock.Lock()
	//先通过pid从世界管理中获取player对象
	player := wm.Players[pid]
	//从世界地图中删除
	wm.AoiMgr.RemoveFromGridByPos(int(pid), player.X, player.Z)
	//从世界管理器中删除
	delete(wm.Players, pid)
	wm.PLock.Unlock()
}

//GetPlayerByPid 通过玩家ID获取一个Player对象
func (wm *WorldManager) GetPlayerByPid(pid int32) *Player {
	wm.PLock.RLock()
	p := wm.Players[pid]
	wm.PLock.RUnlock()
	return p
}

//GetAllPlayers 获取全部在线玩家切片的方法
func (wm *WorldManager) GetAllPlayers() []*Player {
	wm.PLock.RLock()
	defer wm.PLock.RUnlock()

	players := make([]*Player, 0)

	//将世界管理器中的Player对象加入到切片中
	for _, player := range wm.Players {
		players = append(players, player)
	}
	//返回全部在线玩家
	return players
}
