package client

import (
	"fmt"
	"micro/demo/net/server"
	"testing"
	"time"
)

func TestConnect(t *testing.T) {
	go func() {
		err := server.Serve(":8080")
		t.Log(err)
	}()
	time.Sleep(3 * time.Second)

	err := Connect("localhost:8080")
	t.Log(err)
}

func TestClient_Connect(t *testing.T) {
	// 启动server
	svc := server.InitServer("tcp", ":8080")
	go func() {
		err := svc.Start()
		t.Log(err)
	}()
	time.Sleep(3 * time.Second)

	client := InitClient("tcp", ":8080")
	str, err := client.Connect("测试数据")
	if err != nil {
		return
	}
	fmt.Println(str)
}
