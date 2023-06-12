package server

import (
	"context"
	"encoding/json"
	"errors"
	"micro/demo/rpc"
	"micro/rpc/protocol"
	"net"
	"reflect"
)

// numOfLengthBytes 长度字段
const numOfLengthBytes = 8

// Server 服务端
type Server struct {
	network string
	addr    string
	service map[string]reflectionStub
}

// InitServer 初始化服务端
func InitServer(network, addr string) *Server {
	return &Server{
		network: network,
		addr:    addr,
		service: make(map[string]reflectionStub, 16),
	}
}

// RegisterService 服务注册
func (s *Server) RegisterService(service protocol.Service) {
	s.service[service.Name()] = reflectionStub{
		svc:   service,
		value: reflect.ValueOf(service),
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
		// 读取请求消息
		reqBytes, err := rpc.ReadMsg(conn)
		if err != nil {
			return err
		}

		// 还原调用信息
		req := &protocol.Request{}
		err = json.Unmarshal(reqBytes, req)
		if err != nil {
			return err
		}

		// TODO 处理数据
		resp, err := s.Invoke(context.Background(), req)
		if err != nil {
			return err
		}

		// 编码数据
		res := rpc.EncodeMsg(resp.Data)
		_, err = conn.Write(res)
		if err != nil {
			return err
		}
	}
}

func (s *Server) Invoke(ctx context.Context, req *protocol.Request) (*protocol.Response, error) {
	// 根据调用信息，发起业务调用
	service, ok := s.service[req.ServiceName]
	if !ok {
		return nil, errors.New("调用服务不存在")
	}

	// 反射出调用信息 执行调用
	resp, err := service.invoke(ctx, req.MethodName, req.Data)
	if err != nil {
		return nil, err
	}

	return &protocol.Response{
		Data: resp,
	}, nil
}

type reflectionStub struct {
	svc   protocol.Service
	value reflect.Value
}

func (r *reflectionStub) invoke(ctx context.Context, methodName string, data []byte) ([]byte, error) {
	// 反射找到方法，执行调用
	method := r.value.MethodByName(methodName)
	in := make([]reflect.Value, 2)
	in[0] = reflect.ValueOf(context.Background())
	inReq := reflect.New(method.Type().In(1).Elem())
	err := json.Unmarshal(data, inReq.Interface())
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
