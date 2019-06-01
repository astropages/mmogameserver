/*
	地图模块（AOI格子管理）
*/

package core

import "fmt"

//AOIManager 地图类
type AOIManager struct {
	MinX    int           //地图左上角坐标
	MaxX    int           //地图右上角坐标
	CountsX int           //地图X轴方向的格子数
	MinY    int           //地图左下角坐标
	MaxY    int           //地图右下角坐标
	CountsY int           //地图Y轴方向的格子数
	grids   map[int]*Grid //地图内格子的ID（key）集合
}

//GridWidth 获取格子宽度
func (m *AOIManager) GridWidth() int {
	return (m.MaxX - m.MinX) / m.CountsX
}

//GridHeight 获取格子高度
func (m *AOIManager) GridHeight() int {
	return (m.MaxY - m.MinY) / m.CountsY
}

//NewAOIManager 初始化地图
func NewAOIManager(minX, maxX, countsX, minY, maxY, countsY int) *AOIManager {
	aoiMgr := &AOIManager{
		MinX:    minX,
		MaxX:    maxX,
		CountsX: countsX,
		MinY:    minY,
		MaxY:    maxY,
		CountsY: countsY,
		grids:   make(map[int]*Grid),
	}

	//初始化地图里的所有格子
	for y := 0; y < countsY; y++ {
		for x := 0; x < countsX; x++ {
			//格子ID
			gid := y*countsX + x
			//添加一个格子
			aoiMgr.grids[gid] = NewGrid(gid,
				aoiMgr.MinX+x*aoiMgr.GridWidth(),
				aoiMgr.MinX+(x+1)*aoiMgr.GridWidth(),
				aoiMgr.MinY+y*aoiMgr.GridHeight(),
				aoiMgr.MinY+(y+1)*aoiMgr.GridHeight())
		}
	}

	return aoiMgr
}

//String 打印地图信息
func (m *AOIManager) String() string {
	s := fmt.Sprintf("AOIManager:\nMinX:%d, MaxX:%d, countsX:%d, minY:%d, maxY:%d, countsY:%d, grids:\n",
		m.MinX, m.MaxX, m.CountsX, m.MinY, m.MaxY, m.CountsY)

	//打印全部格子信息
	for _, grid := range m.grids {
		s += fmt.Sprintln(grid)
	}
	return s
}

//AddPidToGrid 添加一个PlayerID到一个AOI格子中
func (m *AOIManager) AddPidToGrid(pID, gID int) {
	m.grids[gID].Add(pID, nil)
}

//RemovePidFromGrid 从一个AOI格子中移除一个PlayerID
func (m *AOIManager) RemovePidFromGrid(pID, gID int) {
	m.grids[gID].Remove(pID)
}

//GetPidsByGrid 通过格子ID获取当前格子的全部PlayerID
func (m *AOIManager) GetPidsByGrid(gID int) (playerIDs []int) {
	playerIDs = m.grids[gID].GetPlayerIDs()
	return
}

//GetSurroundGridsByGid 通过一个格子ID获取当前格子周边的九宫格范围内的所有格子ID集合，返回的grids是一个九宫格切片
func (m *AOIManager) GetSurroundGridsByGid(gID int) (grids []*Grid) {
	//判断gID是否在AOI中
	if _, ok := m.grids[gID]; !ok {
		return
	}

	//将当前gID（九宫格中心）放入九宫格切片中
	grids = append(grids, m.grids[gID])

	//分别判断gID左边和右边是否有格子

	idx := gID % m.CountsX //通过id获取当前gID的X轴编号

	if idx > 0 { //判断当前gID左边是否有格子
		grids = append(grids, m.grids[gID-1]) //将格子加入到九宫格切片中
	}
	if idx < m.CountsX-1 { //判断当前gID右边是否有格子
		grids = append(grids, m.grids[gID+1]) //将格子加入到九宫格切片中
	}

	//依次判断此集合格子的上下是否有格子

	gidsX := make([]int, 0, len(grids)) //将X轴的全部Grid的ID放到一个切片
	for _, v := range grids {           //遍历添加到切片
		gidsX = append(gidsX, v.GID)
	}
	for _, gid := range gidsX {
		idy := gid / m.CountsX //通过id获取当前gID的Y轴编号
		if idy > 0 {           //判断当前gID上方是否有格子
			grids = append(grids, m.grids[gid-m.CountsX]) //将格子加入到九宫格切片中
		}
		if idy < m.CountsY-1 { //判断当前gID下方是否有格子
			grids = append(grids, m.grids[gid+m.CountsX]) //将格子加入到九宫格切片中
		}
	}

	return
}

//GetGridByPos 根据坐标获取当前格子的ID
func (m *AOIManager) GetGridByPos(x, y float32) int {
	if x < 0 || int(x) >= m.MaxX {
		return -1
	}
	if y < 0 || int(y) >= m.MaxY {
		return -1
	}
	//根据坐标获取当前格子的ID
	idx := (int(x) - m.MinX) / m.GridWidth()
	idy := (int(y) - m.MinY) / m.GridHeight()

	gid := idy*m.CountsX + idx

	return gid
}

//GetSurroundPIDsByPos 根据坐标获取周边九宫格范围内的全部玩家ID
func (m *AOIManager) GetSurroundPIDsByPos(x, y float32) (playerIDs []int) {

	//根据坐标获取当前所在格子ID
	gid := m.GetGridByPos(x, y)
	fmt.Printf("Grid:%d\n", gid)

	//通过格子ID获取周边九宫格范围内的格子集合
	grids := m.GetSurroundGridsByGid(gid)

	//将九宫格内范围内的所有玩家ID放入一个切片中
	for _, grid := range grids {
		playerIDs = append(playerIDs, grid.GetPlayerIDs()...)
	}

	return
}

//AddToGridByPos 通过坐标将一个pID加入到格子中
func (m *AOIManager) AddToGridByPos(pID int, x, y float32) {
	gID := m.GetGridByPos(x, y)
	//获取当前格子
	grid := m.grids[gID]
	//向格子添加玩家
	grid.Add(pID, nil)
}

//RemoveFromGridByPos 通过坐标将一个player从格子中删除
func (m *AOIManager) RemoveFromGridByPos(pID int, x, y float32) {
	gID := m.GetGridByPos(x, y)

	grid := m.grids[gID]

	grid.Remove(pID)
}
