package tracing

import (
	"fmt"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"net"
)

const instrumentationName = "github.com/Tini-Bytes/micro/tracing"

type ServerTracingBuilder struct {
	Tracer trace.Tracer
	Port   int
}

func (b *ServerTracingBuilder) Build() grpc.UnaryServerInterceptor {
	if b.Tracer == nil {
		b.Tracer = otel.GetTracerProvider().Tracer(instrumentationName)
	}

	addr := GetOutboundIP()
	if b.Port != 0 {
		addr = fmt.Sprintf("%s:%d", addr, b.Port)
	}

	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		ctx = b.extract(ctx)
		spanCtx, span := b.Tracer.Start(ctx, info.FullMethod, trace.WithSpanKind(trace.SpanKindServer))
		span.SetAttributes(attribute.String("addr", addr))

		defer func() {
			// recode error
			if err != nil {
				span.SetStatus(codes.Error, err.Error())
				span.RecordError(err)
			}
			span.End()
		}()

		resp, err = handler(spanCtx, req)
		return
	}
}

func (b *ServerTracingBuilder) extract(ctx context.Context) context.Context {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		md = metadata.MD{}
	}
	return otel.GetTextMapPropagator().Extract(ctx, propagation.HeaderCarrier(md))
}

func GetOutboundIP() string {
	// ping DNS
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return ""
	}
	defer func() {
		_ = conn.Close()
	}()
	addr := conn.LocalAddr().(*net.UDPAddr)
	return addr.String()
}
