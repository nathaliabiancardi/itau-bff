package observability

import "github.com/prometheus/client_golang/prometheus"

var (
	RequestCount = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total de requests",
		},
		[]string{"path", "status"},
	)

	RequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Latência das requests",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"path"},
	)
)

func RegisterMetrics() {
	prometheus.MustRegister(RequestCount)
	prometheus.MustRegister(RequestDuration)
}
