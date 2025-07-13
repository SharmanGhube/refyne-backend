package metrics

import "github.com/prometheus/client_golang/prometheus"

var (
	HttpRequests = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "refyne",
			Subsystem: "http",
			Name:      "requests_total",
			Help:      "Total number of HTTP requests",
		},
		[]string{"method", "path", "status", "route"},
	)

	RequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Duration of HTTP requests",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path"},
	)
)

func Register() {
	prometheus.MustRegister(HttpRequests, RequestDuration)
}
