package pool

import (
	"context"
	"errors"
	"net"
	"sync"
	"time"
)

// idleConn 空闲连接
type idleConn struct {
	c              net.Conn
	lastActiveTime time.Time
}

// connReq 请求队列
type connReq struct {
	connChan chan net.Conn
}

// Pool 连接池
type Pool struct {
	idlesConn   chan *idleConn           //空闲连接队列
	reqsConn    []connReq                // 请求队列
	maxActive   int                      // 最大连接数
	curActive   int                      // 当前连接数
	maxIdleTime time.Duration            // 最大空闲时间
	factory     func() (net.Conn, error) // 连接接口
	lock        sync.RWMutex
}

func NewPool(initCap, maxActive, maxIdle int, maxIdleTime time.Duration, factory func() (net.Conn, error)) (*Pool, error) {
	// 参数较验
	if initCap > maxIdle {
		return nil, errors.New("micro: 初始连接数量不能大于最大空闲连接数")
	}

	idle := make(chan *idleConn, maxIdle)
	// 初始化连接
	for i := 0; i < initCap; i++ {
		conn, err := factory()
		if err != nil {
			return nil, err
		}

		// 将新建的conn放入空闲队列
		idle <- &idleConn{
			c:              conn,
			lastActiveTime: time.Now(),
		}
	}

	return &Pool{
		idlesConn:   idle,
		reqsConn:    nil,
		maxActive:   maxActive,
		curActive:   0,
		maxIdleTime: maxIdleTime,
		factory:     factory,
	}, nil
}

// Get 获取
func (p *Pool) Get(ctx context.Context) (net.Conn, error) {
	select {
	case <-ctx.Done():
		// 超时
		return nil, ctx.Err()
	default:

	}

	// 从空闲队列拿
	for {
		select {
		case ic := <-p.idlesConn:
			// 拿到了空闲连接

			if ic.lastActiveTime.Add(p.maxIdleTime).Before(time.Now()) {
				// 加上最大空闲时间都小于当前时间，说明过期
				_ = ic.c.Close()
				continue
			}
			// 返回连接
			return ic.c, nil
		default:
			// 没有空闲连接
			p.lock.Lock()
			if p.curActive >= p.maxActive {
				// 超过连接上限 -> 进入请求队列
				req := connReq{
					connChan: make(chan net.Conn, 1),
				}
				p.reqsConn = append(p.reqsConn, req)
				p.lock.Unlock()

				// 等待归还
				select {
				case <-ctx.Done():
					// 超时 -> 转发
					go func() {
						c := <-req.connChan
						_ = p.Put(context.Background(), c)
					}()
					return nil, ctx.Err()
				case c := <-req.connChan:
					return c, nil
				}
			}

			// 没超出上限 -> 新建连接
			c, err := p.factory()
			if err != nil {
				return nil, err
			}
			p.curActive++
			p.lock.Unlock()
			return c, nil
		}
	}
}

func (p *Pool) Put(ctx context.Context, c net.Conn) error {
	p.lock.Lock()
	if len(p.reqsConn) > 0 {
		// 有阻塞的请求 -> 在阻塞队列拿出一个请求
		req := p.reqsConn[0]
		p.reqsConn = p.reqsConn[1:]
		p.lock.Unlock()
		req.connChan <- c
		return nil
	}

	// 没有阻塞的请求 -> 放入空闲队列
	p.lock.Unlock()
	ic := &idleConn{
		c:              c,
		lastActiveTime: time.Now(),
	}
	select {
	case p.idlesConn <- ic:
		// 空闲队列没慢 -> 直接放入chan
	default:
		// 空闲队列满了 -> 关闭
		_ = c.Close()
		p.lock.Lock()
		p.curActive--
		p.lock.Unlock()
	}
	return nil
}
