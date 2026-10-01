package middleware

import (
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	httpRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "video_feed_http_requests_total",
			Help: "HTTP 请求总数",
		},
		[]string{"method", "route", "status"},
	)

	httpRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "video_feed_http_request_duration_seconds",
			Help:    "HTTP 请求处理耗时",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "route", "status"},
	)

	httpRequestsInFlight = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "video_feed_http_requests_in_flight",
			Help: "当前正在处理的 HTTP 请求数",
		},
	)
)

func init() {
	prometheus.MustRegister(
		httpRequestsTotal,
		httpRequestDuration,
		httpRequestsInFlight,
	)
}

func PrometheusMetrics() gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.Request.URL.Path
		if path == "/metrics" || path == "/ping" || strings.HasPrefix(path, "/uploads/") {
			c.Next()
			return
		}

		start := time.Now()

		httpRequestsInFlight.Inc()
		defer httpRequestsInFlight.Dec()

		c.Next()

		route := c.FullPath()
		if route == "" {
			route = "unmatched"
		}

		method := c.Request.Method
		status := strconv.Itoa(c.Writer.Status())

		httpRequestsTotal.
			WithLabelValues(method, route, status).
			Inc()

		httpRequestDuration.
			WithLabelValues(method, route, status).
			Observe(time.Since(start).Seconds())
	}
}
