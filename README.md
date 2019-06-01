# mmogameserver
A demo of MMO game server using tinyserver



Demo Client:

* clientRobot.go



*clientRobot.go*

```go
/*
	模拟游戏客户端
*/

package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"math/rand"
	"mmogameserver/pb"
	"net"
	"os"
	"os/signal"
	"time"

	"github.com/golang/protobuf/proto"
)

//TCPClient 客户端类
type TCPClient struct {
	conn net.Conn //通信socket
	Pid  int32    //玩家ID
	X    float32
	Y    float32
	Z    float32
	V    float32
}

//Message 数据包类
type Message struct {
	MsgID uint32 //消息ID
	Len   uint32 //数据长度
	Data  []byte //数据内容
}

//NewTCPClient 初始化客户端对象
func NewTCPClient(ip string, port int) *TCPClient {
	addrStr := fmt.Sprintf("%s:%d", ip, port)
	conn, err := net.Dial("tcp", addrStr)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	//请求成功
	client := &TCPClient{
		conn: conn,
		Pid:  0,
		X:    0,
		Y:    0,
		Z:    0,
		V:    0,
	}
	fmt.Println("服务器连接成功")
	return client
}

//Unpack 数据拆包方法
func (t *TCPClient) Unpack(headData []byte) (*Message, error) {

	headBufReader := bytes.NewReader(headData)
	head := &Message{}

	//读取Len
	if err := binary.Read(headBufReader, binary.LittleEndian, &head.Len); err != nil {
		return nil, err
	}
	//读取MsgID
	if err := binary.Read(headBufReader, binary.LittleEndian, &head.MsgID); err != nil {
		return nil, err
	}

	return head, nil
}

//Pack 数据封包方法
func (t *TCPClient) Pack(msgid uint32, data []byte) ([]byte, error) {
	outbuuff := bytes.NewBuffer([]byte{})
	//写入数据长度
	if err := binary.Write(outbuuff, binary.LittleEndian, uint32(len(data))); err != nil {
		fmt.Println(err)
		return nil, err
	}
	//写入消息ID
	if err := binary.Write(outbuuff, binary.LittleEndian, msgid); err != nil {
		fmt.Println(err)
		return nil, err
	}
	//写入数据内容
	if err := binary.Write(outbuuff, binary.LittleEndian, data); err != nil {
		fmt.Println(err)
		return nil, err
	}
	return outbuuff.Bytes(), nil
}

//SendMsg 当前客户端发送数据包的方法
func (t *TCPClient) SendMsg(msgID uint32, data proto.Message) {
	//打包二进制数据
	binaryData, err := proto.Marshal(data)
	if err != nil {
		fmt.Println(err)
		return
	}
	//打包服务器框架的协议数据
	sendData, err := t.Pack(msgID, binaryData)
	if err == nil {
		//发送给对端
		t.conn.Write(sendData)
	} else {
		fmt.Println(err)
	}
}

//RobotAction 机器人聊天和移动动作的方法
func (t *TCPClient) RobotAction() {
	tp := rand.Intn(2)
	if tp == 0 {
		//自动聊天
		content := fmt.Sprintf("你好，我是player%d", t.Pid)
		msg := &pb.Talk{
			Content: content,
		}
		//将数据发送给服务端
		t.SendMsg(2, msg)
	} else {
		//自动移动
		x := t.X
		z := t.Z

		randPos := rand.Intn(2)
		if randPos == 0 {
			x -= float32(rand.Intn(10))
			z -= float32(rand.Intn(10))
		} else {
			x += float32(rand.Intn(10))
			z += float32(rand.Intn(10))
		}

		//坐标纠正
		if x > 410 {
			x = 410
		} else if x < 85 {
			x = 85
		}

		if z > 400 {
			z = 400
		} else if z < 75 {
			z = 75
		}

		randV := rand.Intn(2)
		v := t.V
		if randV == 0 {
			v = 25
		} else {
			v = 350
		}

		//定义一个proto消息
		msg := &pb.Position{
			X: x,
			Y: t.Y,
			Z: z,
			V: v,
		}
		fmt.Printf("玩家%d移动\n", t.Pid)
		t.SendMsg(3, msg)
	}
}

//DoMsg 根据服务器回执的不同消息处理不同业务的方法
func (t *TCPClient) DoMsg(msg *Message) {

	if msg.MsgID == 1 {
		//服务器回执：分配ID
		//解析proto协议
		syncpid := &pb.SyncPid{}
		proto.Unmarshal(msg.Data, syncpid)

		//给客户端对象赋值
		t.Pid = syncpid.Pid

	} else if msg.MsgID == 200 {
		//服务器回执：广播数据
		bdata := &pb.BroadCast{}
		proto.Unmarshal(msg.Data, bdata)
		if bdata.Tp == 2 && bdata.Pid == t.Pid {
			//服务器给客户端分配了初始化位置坐标
			//更新当前客户端玩家坐标
			t.X = bdata.GetP().X
			t.Y = bdata.GetP().Y
			t.Z = bdata.GetP().Z
			t.V = bdata.GetP().V
			fmt.Printf("玩家%d上线，位置：%f,%f,%f,%f\n", t.Pid, t.X, t.Y, t.Z, t.V)

			//客户端主动请求动作
			go func() {
				for {
					t.RobotAction() //自动完成一个机器人动作
					time.Sleep(5 * time.Second)
				}
			}()
		} else if bdata.Tp == 1 {
			//世界聊天广播的消息
			fmt.Printf("世界聊天 | 玩家%d: %s\n", bdata.Pid, bdata.GetContent())
		}

	}
}

//Start 客户端读写业务的方法
func (t *TCPClient) Start() {
	go func() {
		for {

			//按照服务器框架的数据协议先获取数据包的head部分
			headData := make([]byte, 8)
			if _, err := io.ReadFull(t.conn, headData); err != nil {
				fmt.Println(err)
				return
			}
			messageHead, err := t.Unpack(headData)
			if err != nil {
				return
			}
			//根据数据长度获取数据包的Data部分
			if messageHead.Len > 0 {
				messageHead.Data = make([]byte, messageHead.Len)
				if _, err := io.ReadFull(t.conn, messageHead.Data); err != nil {
					fmt.Println(err)
					return
				}
			}
			//根据不同的MsgID来处理不同的业务
			t.DoMsg(messageHead)
			time.Sleep(3 * time.Second)
		}
	}()

}

func main() {

	//模拟上线2个机器人玩家
	for i := 0; i < 2; i++ {
		//初始化客户端对象
		client := NewTCPClient("127.0.0.1", 8999)

		//客户端读写业务
		client.Start()

		time.Sleep(1 * time.Second)
	}

	//通过系统信号阻塞
	c := make(chan os.Signal, 1)            //接受系统信息的通道
	signal.Notify(c, os.Kill, os.Interrupt) //捕获信号到通道
	sig := <-c                              //阻塞接收信号
	fmt.Println("系统信号：", sig)
	return

}

```

