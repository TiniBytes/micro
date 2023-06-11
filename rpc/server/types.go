package server

type Service interface {
	Name() string
}

type Request struct {
	ServiceName string
	MethodName  string
	Arg         []byte
}

type Response struct {
	Data []byte
}
