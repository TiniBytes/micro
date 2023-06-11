package server

import (
	"encoding/binary"
	"fmt"
	"net"
)

// numOfLengthBytes 长度字段
const numOfLengthBytes = 8

// Server 服务端
type Server struct {
	network string
	addr    string
}

// InitServer 初始化服务端
func InitServer(network, addr string) *Server {
	return &Server{
		network: network,
		addr:    addr,
	}
}

// Start 服务器启动
func (s *Server) Start() error {
	// 监听端口
	listener, err := net.Listen(s.network, s.addr)
	if err != nil {
		return err
	}

	for {
		// 接收连接
		conn, err := listener.Accept()
		if err != nil {
			return err
		}

		// 处理请求
		go func() {
			err := s.handleConn(conn)
			if err != nil {
				_ = conn.Close()
			}
		}()
	}
}

// handleConn 处理连接
// 长度字段：8字节
// 请求数据：根据长度字段确定
func (s *Server) handleConn(conn net.Conn) error {
	for {
		// 读数据：读数据根据上层协议决定怎么读
		// 例如，简单的RPC协议一般是分成两段读，先读头部，
		// 根据头部得知Body有多长，再把剩下的数据集读出来
		byteLen := make([]byte, numOfLengthBytes)
		_, err := conn.Read(byteLen)
		if err != nil {
			return err
		}
		// 根据消息长度读数据
		length := binary.BigEndian.Uint64(byteLen)
		reqBytes := make([]byte, length)
		_, err = conn.Read(reqBytes)
		if err != nil {
			return err
		}

		// TODO 处理数据
		rspData := handleMsg(reqBytes)

		// 写回响应：即使处理数据错误，也要返回错误给客户端
		// 不然客户端不知道处理出错
		// data = rspLen的64位表示 + rspData
		rspLen := len(rspData)
		res := make([]byte, numOfLengthBytes+rspLen)

		// 写入长度 + 数据
		binary.BigEndian.PutUint64(res[:numOfLengthBytes], uint64(rspLen))
		copy(res[numOfLengthBytes:], rspData)

		_, err = conn.Write(res)
		if err != nil {
			return err
		}
	}
}

// Serve 服务端
func Serve(addr string) error {
	// 监听端口
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	for {
		// 接收连接
		conn, err := listener.Accept()
		if err != nil {
			return err
		}

		// 处理请求
		go func() {
			err := handleConn(conn)
			if err != nil {
				_ = conn.Close()
			}
		}()
	}
}

// handleConn 处理连接
func handleConn(conn net.Conn) error {
	for {
		// 读数据：读数据根据上层协议决定怎么读
		// 例如，简单的RPC协议一般是分成两段读，先读头部，
		// 根据头部得知Body有多长，再把剩下的数据集读出来
		bs := make([]byte, 8)
		_, err := conn.Read(bs)
		if err != nil {
			return err
		}

		// TODO 处理数据
		msg := handleMsg(bs)

		// 写回响应：即使处理数据错误，也要返回错误给客户端
		// 不然客户端不知道处理出错
		_, err = conn.Write(msg)
		if err != nil {
			return err
		}
	}
}

// handleMsg 处理数据
func handleMsg(req []byte) []byte {
	// TODO 业务逻辑
	fmt.Println("处理数据")
	return req
}
