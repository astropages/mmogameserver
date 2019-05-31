/*
	格子模块（AOI兴趣点）
*/

package core

import (
	"fmt"
	"sync"
)

//Grid 格子类
type Grid struct {
	GID       int                 //格子ID
	MinX      int                 //格子左上角坐标
	MaxX      int                 //格子右上角坐标
	MinY      int                 //格子左下角坐标
	MaxY      int                 //格子右下角坐标
	playerIDs map[int]interface{} //当前格子内玩家或物体的ID（key）集合
	pIDLock   sync.RWMutex        //当前格子内侧集合内容锁
}

//NewGrid 初始化格子
func NewGrid(gID, minX, maxX, minY, maxY int) *Grid {
	return &Grid{
		GID:       gID,
		MinX:      minX,
		MaxX:      maxX,
		MinY:      minY,
		MaxY:      maxY,
		playerIDs: make(map[int]interface{}),
	}
}

//Add 给格子添加一个玩家或物体的方法
func (g *Grid) Add(playerID int, player interface{}) {
	g.pIDLock.Lock()
	defer g.pIDLock.Unlock()

	g.playerIDs[playerID] = player
}

//Remove 从格子中删除一个玩家或物体的方法
func (g *Grid) Remove(playerID int) {
	g.pIDLock.Lock()
	defer g.pIDLock.Unlock()

	delete(g.playerIDs, playerID)
}

//GetPlayerIDs 获取当前格子所有玩家或物体ID的方法
func (g *Grid) GetPlayerIDs() (playerIDs []int) {
	g.pIDLock.RLock()
	defer g.pIDLock.RUnlock()

	//将集合中的所有key保存到切片
	for playerID := range g.playerIDs {
		playerIDs = append(playerIDs, playerID)
	}

	return
}

//调试打印格子信息的方法
func (g *Grid) String() string {
	return fmt.Sprintf("Grid:%d, minX:%d, maxX:%d, minY:%d, maxY:%d, playerIDs:%v",
		g.GID, g.MinX, g.MaxX, g.MinX, g.MaxY, g.playerIDs)
}
