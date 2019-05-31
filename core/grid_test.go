package core

import (
	"fmt"
	"testing"
)

func TestGrid(t *testing.T) {
	player1 := "玩家一"
	player2 := "玩家二"

	//单元测试Grid模块
	g := NewGrid(1, 2, 3, 10, 20)

	g.Add(1, player1)
	g.Add(2, player2)

	//打出格子信息
	fmt.Println(g)
}
