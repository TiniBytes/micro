package server

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"errors"
	"net"
	"reflect"
)

// numOfLengthBytes 长度字段
const numOfLengthBytes = 8

// Server 服务端
type Server struct {
	network string
	addr    string
	service map[string]Service
}

// InitServer 初始化服务端
func InitServer(network, addr string) *Server {
	return &Server{
		network: network,
		addr:    addr,
		service: make(map[string]Service, 16),
	}
}

// RegisterService 服务注册
func (s *Server) RegisterService(service Service) {
	s.service[service.Name()] = service
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
		rspData, err := s.handleMsg(reqBytes)
		if err != nil {
			// 业务err, 暂时忽略
			return err
		}

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

// handleMsg 处理数据
func (s *Server) handleMsg(reqData []byte) ([]byte, error) {
	// 还原调用信息
	req := &Request{}
	err := json.Unmarshal(reqData, req)
	if err != nil {
		return nil, err
	}

	// 根据调用信息，发起业务调用
	service, ok := s.service[req.ServiceName]
	if !ok {
		return nil, errors.New("调用服务不存在")
	}

	// 反射找到方法，执行调用
	val := reflect.ValueOf(service)
	method := val.MethodByName(req.MethodName)
	in := make([]reflect.Value, 2)
	in[0] = reflect.ValueOf(context.Background())
	inReq := reflect.New(method.Type().In(1).Elem())
	err = json.Unmarshal(req.Arg, inReq.Interface())
	if err != nil {
		return nil, err
	}
	in[1] = inReq
	// result[0]: 返回值    result[1]: error
	results := method.Call(in)
	if results[1].Interface() != nil {
		return nil, results[1].Interface().(error)
	}

	// 返回信息
	return json.Marshal(results[0].Interface())
}
