package client

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/silenceper/pool"
	"micro/demo/rpc"
	"net"
	"reflect"
	"time"
)

// InitClientProxy 初始化客户端代理
// 为函数类型字段赋值
func InitClientProxy(addr string, service Service) error {
	// 初始化proxy
	client, err := NewClient(addr)
	if err != nil {
		return err
	}

	return setFuncField(service, client)
}

// setFuncField 捕捉本地调用
func setFuncField(service Service, p Proxy) error {
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

				reqData, err := json.Marshal(args[1].Interface())
				if err != nil {
					return []reflect.Value{retVal, reflect.ValueOf(err)}
				}
				req := &Request{
					ServiceName: service.Name(),
					MethodName:  fieldTyp.Name,
					Arg:         reqData,
				}

				// 发起RPC调用
				resp, err := p.Invoke(ctx, req)
				if err != nil {
					return []reflect.Value{retVal, reflect.ValueOf(err)}
				}
				err = json.Unmarshal(resp.Data, retVal.Interface())
				if err != nil {
					return []reflect.Value{retVal, reflect.ValueOf(err)}
				}

				return []reflect.Value{retVal, reflect.Zero(reflect.TypeOf(new(error)).Elem())}
			}
			// 设置值
			fnVal := reflect.MakeFunc(fieldTyp.Type, fn)
			fieldVal.Set(fnVal)
		}

	}
	return nil
}

const numOfLengthBytes = 8

type Client struct {
	addr string
	pool pool.Pool
}

func NewClient(addr string) (*Client, error) {
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

	return &Client{
		addr: addr,
		pool: p,
	}, nil
}

// Invoke 发送请求到服务端
func (c *Client) Invoke(ctx context.Context, req *Request) (*Response, error) {
	data, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	// 从连接池拿一个连接
	val, err := c.pool.Get()
	if err != nil {
		return nil, err
	}
	conn := val.(net.Conn)
	resp, err := c.Send(conn, data)
	if err != nil {
		return nil, err
	}

	return &Response{
		Data: resp,
	}, nil
}

// Send 向服务端发送数据
func (c *Client) Send(conn net.Conn, data []byte) ([]byte, error) {
	// 写入数据
	req := rpc.EncodeMsg(data)
	_, err := conn.Write(req)
	if err != nil {
		return nil, err
	}

	// 接收响应数据
	return rpc.ReadMsg(conn)
}
