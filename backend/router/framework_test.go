package router

import (
	"context"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"video_feedsystem/middleware"
	"video_feedsystem/storage"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/cloudwego/hertz/pkg/common/ut"
	"github.com/cloudwego/hertz/pkg/network/standard"
	"github.com/hertz-contrib/cors"
)

// 不连接数据库和 Redis，检查框架切换后的路由、鉴权边界和媒体协议。
func TestHertzHTTPCompatibility(t *testing.T) {
	if err := storage.Init(t.TempDir()); err != nil {
		t.Fatal(err)
	}
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()
	h := server.New(server.WithListener(listener), server.WithTransport(standard.NewTransporter))
	h.Use(middleware.PrometheusMetrics())
	h.Use(cors.New(cors.Config{
		AllowOrigins: []string{"http://localhost:5173"},
		AllowMethods: []string{"GET", "POST", "DELETE", "OPTIONS"},
		AllowHeaders: []string{"Content-Type", "Authorization"},
	}))
	InitRouter(h, true)

	protected := []struct{ method, path string }{
		{"POST", "/account/logout"},
		{"GET", "/feed/following"},
		{"POST", "/video/upload-video"},
		{"POST", "/video/upload-cover"},
		{"POST", "/video/publish"},
		{"POST", "/video/like"},
		{"POST", "/video/unlike"},
		{"GET", "/video/like-status"},
		{"GET", "/video/liked"},
		{"POST", "/comment/publish"},
		{"DELETE", "/comment/delete"},
		{"POST", "/social/follow"},
		{"POST", "/social/unfollow"},
		{"GET", "/social/status"},
		{"GET", "/social/following"},
		{"GET", "/social/followers"},
	}
	for _, route := range protected {
		t.Run(route.method+route.path, func(t *testing.T) {
			response := ut.PerformRequest(h.Engine, route.method, route.path, nil)
			if response.Code != 401 || !strings.Contains(response.Body.String(), "请求未携带 Token") {
				t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
			}
		})
	}
	for _, path := range []string{
		"/feed/list?cursor=-1", "/video/detail?video_id=invalid",
		"/video/list-by-author-id?author_id=invalid", "/account/profile?account_id=invalid",
		"/account/search?keyword=", "/comment/list?video_id=invalid",
	} {
		response := ut.PerformRequest(h.Engine, "GET", path, nil)
		if response.Code != 400 || !json.Valid(response.Body.Bytes()) {
			t.Fatalf("%s: status=%d body=%s", path, response.Code, response.Body.String())
		}
	}
	for _, path := range []string{"/account/register", "/account/login", "/account/refresh"} {
		response := ut.PerformRequest(h.Engine, "POST", path,
			&ut.Body{Body: strings.NewReader("{"), Len: 1},
			ut.Header{Key: "Content-Type", Value: "application/json"})
		if response.Code != 400 || !strings.Contains(response.Body.String(), "JSON 解析失败") {
			t.Fatalf("%s: status=%d body=%s", path, response.Code, response.Body.String())
		}
	}
	response := ut.PerformRequest(h.Engine, "GET", "/ping", nil)
	if response.Code != 200 || !strings.Contains(response.Body.String(), "pong") {
		t.Fatalf("ping: %d %s", response.Code, response.Body.String())
	}
	// net/http 适配器使用连接写响应，所以 /metrics 走真实本地 HTTP 连接验证。
	runResult := make(chan error, 1)
	go func() { runResult <- h.Run() }()
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		if err := h.Shutdown(ctx); err != nil {
			t.Error(err)
		}
		if err := <-runResult; err != nil {
			t.Error(err)
		}
	}()
	client := &http.Client{Timeout: 5 * time.Second}
	metricsResponse, err := client.Get("http://" + listener.Addr().String() + "/metrics")
	if err != nil {
		t.Fatal(err)
	}
	metricsBody, err := io.ReadAll(metricsResponse.Body)
	metricsResponse.Body.Close()
	client.CloseIdleConnections()
	if err != nil || metricsResponse.StatusCode != 200 || !strings.Contains(string(metricsBody), "video_feed_http_request_duration_seconds_bucket") {
		t.Fatalf("metrics: status=%d error=%v", metricsResponse.StatusCode, err)
	}
	response = ut.PerformRequest(h.Engine, "OPTIONS", "/video/publish", nil,
		ut.Header{Key: "Origin", Value: "http://localhost:5173"},
		ut.Header{Key: "Access-Control-Request-Method", Value: "POST"},
		ut.Header{Key: "Access-Control-Request-Headers", Value: "authorization,content-type"})
	if response.Code != 204 || string(response.Header().Peek("Access-Control-Allow-Origin")) != "http://localhost:5173" {
		t.Fatalf("CORS preflight: status=%d", response.Code)
	}
	uploads := filepath.Join(storage.Root(), "uploads")
	if err := os.MkdirAll(uploads, 0755); err != nil {
		t.Fatal(err)
	}
	mediaFile := filepath.Join(uploads, "range.mp4")
	if err := os.WriteFile(mediaFile, []byte("0123456789"), 0600); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		// Hertz 会短时缓存文件句柄，Windows 需要等句柄释放后才能清理测试文件。
		deadline := time.Now().Add(20 * time.Second)
		for {
			err := os.Remove(mediaFile)
			if err == nil || os.IsNotExist(err) {
				return
			}
			if time.Now().After(deadline) {
				t.Error(err)
				return
			}
			time.Sleep(100 * time.Millisecond)
		}
	})
	response = ut.PerformRequest(h.Engine, "GET", "/uploads/range.mp4", nil,
		ut.Header{Key: "Range", Value: "bytes=2-5"})
	if response.Code != 206 || response.Body.String() != "2345" || string(response.Header().Peek("Content-Range")) != "bytes 2-5/10" {
		t.Fatalf("Range: status=%d body=%s", response.Code, response.Body.String())
	}
	// 上传开关本身不依赖 Redis，单独验证关闭时会阻断后续处理。
	blocked := server.New()
	blocked.POST("/publish", middleware.RequireUploadEnabled(false), func(_ context.Context, c *app.RequestContext) {
		t.Error("上传关闭后不应继续处理")
	})
	response = ut.PerformRequest(blocked.Engine, "POST", "/publish", nil)
	if response.Code != 403 {
		t.Fatalf("upload disabled: status=%d", response.Code)
	}
}
