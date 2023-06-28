package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	_ "github.com/prometheus/client_golang/prometheus"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"time"
)

type ServerMetricsBuilder struct {
	Namespace string
	Subsystem string
}

func (b *ServerMetricsBuilder) Build() grpc.UnaryServerInterceptor {
	reqCount := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: b.Namespace,
		Subsystem: b.Subsystem,
		Name:      "active-request-count",
		Help:      "current-request-info",
		ConstLabels: map[string]string{
			"component": "server",
		},
	}, []string{"service"})
	prometheus.MustRegister(reqCount)

	response := prometheus.NewSummaryVec(prometheus.SummaryOpts{
		Namespace: b.Namespace,
		Subsystem: b.Subsystem,
		Name:      "response-info",
		Help:      "current-response-info",
		ConstLabels: map[string]string{
			"component": "server",
		},
	}, []string{"service"})
	prometheus.MustRegister(response)

	errCount := prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: b.Namespace,
		Subsystem: b.Subsystem,
		Name:      "response-info",
		Help:      "current-error-info",
		ConstLabels: map[string]string{
			"component": "server",
		},
	}, []string{"service"})
	prometheus.MustRegister(errCount)

	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		startTime := time.Now()

		// Record requests
		reqCount.WithLabelValues(info.FullMethod).Add(1)

		// remove requests
		defer func() {
			reqCount.WithLabelValues(info.FullMethod).Add(-1)
			if err != nil {
				errCount.WithLabelValues(info.FullMethod).Add(1)
			}

			response.WithLabelValues(info.FullMethod).Observe(float64(time.Now().Sub(startTime).Milliseconds()))
		}()

		resp, err = handler(ctx, req)
		return
	}
}
