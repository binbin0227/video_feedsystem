package middleware

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
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

func PrometheusMetrics() app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		path := string(c.Path())
		if path == "/metrics" || path == "/ping" || strings.HasPrefix(path, "/uploads/") {
			c.Next(ctx)
			return
		}

		start := time.Now()

		httpRequestsInFlight.Inc()
		defer httpRequestsInFlight.Dec()

		c.Next(ctx)

		route := c.FullPath()
		if route == "" {
			route = "unmatched"
		}

		method := string(c.Method())
		status := strconv.Itoa(c.Response.StatusCode())

		httpRequestsTotal.
			WithLabelValues(method, route, status).
			Inc()

		httpRequestDuration.
			WithLabelValues(method, route, status).
			Observe(time.Since(start).Seconds())
	}
}
