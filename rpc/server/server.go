package server

import (
	"context"
	"errors"
	"micro/rpc/protocol"
	"micro/rpc/serialize"
	"micro/rpc/serialize/json"
	"net"
	"reflect"
)

// numOfLengthBytes 长度字段
const numOfLengthBytes = 8

// Server 服务端
type Server struct {
	network     string
	addr        string
	service     map[string]reflectionStub
	serializers map[uint8]serialize.Serializer
}

// InitServer 初始化服务端
func InitServer(network, addr string) *Server {
	res := &Server{
		network:     network,
		addr:        addr,
		service:     make(map[string]reflectionStub, 16),
		serializers: make(map[uint8]serialize.Serializer, 4),
	}

	// 注册默认序列化协议
	res.RegisterSerializer(&json.Serializer{})
	return res
}

// RegisterSerializer 注册序列化协议
func (s *Server) RegisterSerializer(serializer serialize.Serializer) {
	// 传入具体实现的序列化协议结构体
	s.serializers[serializer.Code()] = serializer
}

// RegisterService 服务注册
func (s *Server) RegisterService(service protocol.Service) {
	s.service[service.Name()] = reflectionStub{
		svc:         service,
		value:       reflect.ValueOf(service),
		serializers: s.serializers,
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
			err = s.handleConn(conn)
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
		reqBytes, err := protocol.ReadMsg(conn)
		if err != nil {
			return err
		}

		// 还原调用信息
		req := protocol.DecodeRequest(reqBytes)
		if err != nil {
			return err
		}

		// TODO 处理数据
		resp, err := s.Invoke(req)
		if err != nil {
			resp.Error = []byte(err.Error())
		}
		resp.CalculateHeaderLength()
		resp.CalculateBodyLength()

		// 编码数据
		_, err = conn.Write(protocol.EncodeResponse(resp))
		if err != nil {
			return err
		}
	}
}

func (s *Server) Invoke(req *protocol.Request) (*protocol.Response, error) {
	// 根据调用信息，发起业务调用
	service, ok := s.service[req.ServiceName]
	resp := &protocol.Response{
		MessageID:  req.MessageID,
		Version:    req.Version,
		Compress:   req.Compress,
		Serializer: req.Serializer,
	}

	// 捕获错误
	if !ok {
		return resp, errors.New("调用服务不存在")
	}

	// 反射出调用信息 执行调用
	respData, err := service.invoke(req)
	resp.Data = respData
	if err != nil {
		return resp, err
	}

	return resp, nil
}

type reflectionStub struct {
	svc         protocol.Service
	value       reflect.Value
	serializers map[uint8]serialize.Serializer
}

func (r *reflectionStub) invoke(req *protocol.Request) ([]byte, error) {
	// 反射找到方法，执行调用
	method := r.value.MethodByName(req.MethodName)
	in := make([]reflect.Value, 2)
	in[0] = reflect.ValueOf(context.Background())
	inReq := reflect.New(method.Type().In(1).Elem())

	serializer, ok := r.serializers[req.Serializer]
	if !ok {
		return nil, errors.New("micro: 不支持的序列化协议")
	}
	err := serializer.Decode(req.Data, inReq.Interface())
	if err != nil {
		return nil, err
	}
	in[1] = inReq
	// result[0]: 返回值    result[1]: error
	results := method.Call(in)

	if results[1].Interface() != nil {
		err = results[1].Interface().(error)
	}

	var res []byte
	if results[0].IsNil() {
		return nil, err
	} else {
		var er error
		res, er = serializer.Encode(results[0].Interface())
		if er != nil {
			return nil, er
		}
	}
	// 返回信息
	return res, err
}
