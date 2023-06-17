package client

import (
	"context"
	"errors"
	"github.com/silenceper/pool"
	"micro/rpc/protocol"
	"micro/rpc/serialize"
	"micro/rpc/serialize/json"
	"net"
	"reflect"
	"time"
)

// InitClientProxy 初始化客户端代理
// 为函数类型字段赋值
func InitClientProxy(addr string, service protocol.Service, opts ...ClientOption) error {
	// 初始化proxy
	client, err := NewClient(addr, opts...)
	if err != nil {
		return err
	}

	return client.setFuncField(service, client)
}

// setFuncField 捕捉本地调用
func (c *Client) setFuncField(service protocol.Service, p protocol.Proxy) error {
	if service == nil {
		return errors.New("rpc: 不支持nil")
	}

	val := reflect.ValueOf(service)
	typ := val.Type()
	// 只支持指向结构体的一级指针
	if typ.Kind() != reflect.Pointer || typ.Elem().Kind() != reflect.Struct {
		return errors.New("rpc: 只支持指向结构体的一级指针")
	}

	// 捕捉本地调用，用Set方法篡改为RPC调用
	val = val.Elem()
	typ = typ.Elem()
	numField := typ.NumField()
	for i := 0; i < numField; i++ {
		fieldVal := val.Field(i)
		fieldTyp := typ.Field(i)

		if fieldVal.CanSet() {
			// 捕捉本地调用
			fn := func(args []reflect.Value) (results []reflect.Value) {
				retVal := reflect.New(fieldTyp.Type.Out(0).Elem())

				// args[0]: context   args[1]: req
				ctx := args[0].Interface().(context.Context)

				// 序列化
				reqData, err := c.serializer.Encode(args[1].Interface())
				if err != nil {
					return []reflect.Value{retVal, reflect.ValueOf(err)}
				}
				req := &protocol.Request{
					ServiceName: service.Name(),
					MethodName:  fieldTyp.Name,
					Data:        reqData,
					Serializer:  c.serializer.Code(),
				}

				// 发起RPC调用
				req.CalculateHeaderLength()
				req.CalculateBodyLength()
				resp, err := p.Invoke(ctx, req)
				if err != nil {
					return []reflect.Value{retVal, reflect.ValueOf(err)}
				}

				var retErr error
				if len(resp.Error) > 0 {
					retErr = errors.New(string(resp.Error))
				}
				if len(resp.Data) > 0 {
					err = c.serializer.Decode(resp.Data, retVal.Interface())
					if err != nil {
						// 反序列化的err
						return []reflect.Value{retVal, reflect.ValueOf(err)}
					}
				}

				var retErrVal reflect.Value
				if retErr == nil {
					retErrVal = reflect.Zero(reflect.TypeOf(new(error)).Elem())
				} else {
					retErrVal = reflect.ValueOf(retErr)
				}
				return []reflect.Value{retVal, retErrVal}
			}
			// 设置值
			fnVal := reflect.MakeFunc(fieldTyp.Type, fn)
			fieldVal.Set(fnVal)
		}

	}
	return nil
}

type Client struct {
	addr       string
	pool       pool.Pool
	serializer serialize.Serializer
}

type ClientOption func(client *Client)

func ClientWithSerializer(serializer serialize.Serializer) ClientOption {
	return func(client *Client) {
		client.serializer = serializer
	}
}

func NewClient(addr string, opts ...ClientOption) (*Client, error) {
	p, err := pool.NewChannelPool(&pool.Config{
		InitialCap: 1,
		MaxCap:     30,
		MaxIdle:    10,
		Factory: func() (interface{}, error) {
			return net.DialTimeout("tcp", addr, 3*time.Second)
		},
		Close: func(i interface{}) error {
			return i.(net.Conn).Close()
		},
		Ping:        nil,
		IdleTimeout: time.Minute,
	})
	if err != nil {
		return nil, err
	}

	res := &Client{
		addr:       addr,
		pool:       p,
		serializer: &json.Serializer{}, // 默认json序列化
	}

	// 配置选项
	for _, opt := range opts {
		opt(res)
	}
	return res, nil
}

// Invoke 发送请求到服务端
func (c *Client) Invoke(ctx context.Context, req *protocol.Request) (*protocol.Response, error) {
	data := protocol.EncodeRequest(req)

	resp, err := c.Send(data)
	if err != nil {
		return nil, err
	}

	return protocol.DecodeResponse(resp), nil
}

// Send 向服务端发送数据
func (c *Client) Send(data []byte) ([]byte, error) {
	val, err := c.pool.Get()
	if err != nil {
		return nil, err
	}
	conn := val.(net.Conn)
	defer func() {
		_ = conn.Close()
	}()

	// 写入数据
	_, err = conn.Write(data)
	if err != nil {
		return nil, err
	}

	// 接收响应数据
	return protocol.ReadMsg(conn)
}
