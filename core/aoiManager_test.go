package core

import (
	"fmt"
	"testing"
)

//地图测试
func TestAOIManager_init(t *testing.T) {
	//初始化一个AOIManager
	aoiMgr := NewAOIManager(0, 250, 5, 0, 250, 5)

	//打印信息
	fmt.Println(aoiMgr)
}

//九宫格范围测试
func TestAOIManagerSurround(t *testing.T) {
	//初始化一个AOIManager
	aoiMgr := NewAOIManager(0, 250, 5, 0, 250, 5)

	//获取每个格子周边的九宫格信息
	for gid := range aoiMgr.grids {
		grids := aoiMgr.GetSurroundGridsByGid(gid)
		fmt.Printf("gID:%d, grids:%d | ", gid, len(grids))
		//获取以当前gid为中心的九宫格切片
		gIDs := make([]int, 0, len(grids))
		for _, grid := range grids {
			gIDs = append(gIDs, grid.GID)
		}
		fmt.Printf("Surround IDs: %v \n", gIDs)
	}

	fmt.Println("============================")
	//根据坐标获取九宫格范围内的全部玩家ID
	playerIDs := aoiMgr.GetSurroundPIDsByPos(175, 68)
	fmt.Printf("PlayerIDs:%v\n", playerIDs)
	fmt.Println("============================")
}
