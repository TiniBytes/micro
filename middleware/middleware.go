package middleware

import "micro/rpc/protocol"

type HandleFunc func(ctx *protocol.Response)

// Middleware 函数式责任链模式
type Middleware func(next HandleFunc) HandleFunc
