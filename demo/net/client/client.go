package client

import (
	"encoding/binary"
	"fmt"
	"net"
	"time"
)

// numOfLengthBytes 长度字段
const numOfLengthBytes = 8

// Client 客户端
type Client struct {
	network string
	addr    string
}

// InitClient 初始化客户端
func InitClient(network, addr string) *Client {
	return &Client{
		network: network,
		addr:    addr,
	}
}

// Connect  客户端发送数据
func (c *Client) Connect(data string) (string, error) {
	conn, err := net.DialTimeout(c.network, c.addr, 3*time.Second)
	if err != nil {
		return "", err
	}
	defer func() {
		_ = conn.Close()
	}()

	// 发送数据
	str, err := c.Send(conn, data)
	if err != nil {
		return "", err
	}
	return str, conn.Close()
}

// Send 客户端发送数据
func (c *Client) Send(conn net.Conn, data string) (string, error) {
	// 封装数据
	reqLen := len(data)
	req := make([]byte, numOfLengthBytes+reqLen)
	// 写入长度 + 数据
	binary.BigEndian.PutUint64(req[:numOfLengthBytes], uint64(reqLen))
	copy(req[numOfLengthBytes:], data)

	// 发送数据
	_, err := conn.Write(req)
	if err != nil {
		return "", err
	}

	// 接收响应
	byteLen := make([]byte, numOfLengthBytes)
	_, err = conn.Read(byteLen) // 读取长度
	if err != nil {
		return "", err
	}

	// 根据长度读数据
	length := binary.BigEndian.Uint64(byteLen)
	res := make([]byte, length)
	_, err = conn.Read(res) // 根据长度读数据
	if err != nil {
		return "", err
	}

	return string(res), err
}

// Connect 客户端连接
func Connect(addr string) error {
	conn, err := net.DialTimeout("tcp", addr, 3*time.Second)
	if err != nil {
		return err
	}
	defer func() {
		_ = conn.Close()
	}()

	// 发送数据
	err = Send(conn)
	if err != nil {
		return err
	}
	return conn.Close()
}

func Send(conn net.Conn) error {
	// 发送请求
	msg := "hello"
	_, err := conn.Write([]byte(msg))
	if err != nil {
		return err
	}

	// 接收响应
	res := make([]byte, 128)
	_, err = conn.Read(res)
	if err != nil {
		return err
	}

	fmt.Println(res)
	return nil
}
