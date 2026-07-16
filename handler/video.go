package handler

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
	"video_feedsystem/service"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
)

var maxVideoSize int64 = 200 << 20
var maxCoverSize int64 = 10 << 20

type PublishRequest struct {
	Title       string `json:"title" vd:"required;msg:'标题不能为空'"`
	Description string `json:"description"`
	PlayURL     string `json:"play_url" vd:"required;msg:'播放链接不能为空'"`
	CoverURL    string `json:"cover_url" vd:"required;msg:'封面链接不能为空'"`
}

// 发布视频（将信息写入数据库）
func PublishVideo(ctx context.Context, c *app.RequestContext) {
	// 1. 解析 JSON
	var req PublishRequest
	if err := c.BindAndValidate(&req); err != nil {
		c.JSON(consts.StatusBadRequest, map[string]string{"error": "参数错误"})
		return
	}

	// 2. 读取 accountID
	accountIDVal, exists := c.Get("accountID")
	if !exists {
		c.JSON(consts.StatusUnauthorized, map[string]string{"error": "未授权"})
		return
	}
	authorID, ok := accountIDVal.(int64)
	if !ok {
		c.JSON(
			consts.StatusUnauthorized,
			map[string]string{"error": "无效的用户身份"},
		)
		return
	}

	// 3. 存入数据库
	video, err := service.PublishVideo(ctx, authorID, req.Title, req.Description, req.PlayURL, req.CoverURL)
	if err != nil {
		c.JSON(consts.StatusInternalServerError, map[string]string{"error": "视频发布失败，请稍后重试"})
		return
	}

	c.JSON(consts.StatusOK, map[string]any{
		"message": "视频发布成功！",
		"video":   video,
	})

}

// 处理 封面 文件上传到服务器本地硬盘
func UploadCover(ctx context.Context, c *app.RequestContext) {
	// 1. 读取 accountID
	accountIDVal, exists := c.Get("accountID")
	if !exists {
		c.JSON(consts.StatusUnauthorized, map[string]string{"error": "未授权"})
		return
	}
	authorID := accountIDVal.(int64)

	// 2. 接收前端的 "file" 文件
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(consts.StatusBadRequest, map[string]string{"error": "请上传 file 字段'"})
		return
	}

	// 3. 校验后缀名和大小
	ext := strings.ToLower(filepath.Ext(file.Filename))
	// 根据后缀自动归档
	if ext != ".jpg" || ext != ".jpeg" || ext != ".png" {
		c.JSON(consts.StatusBadRequest, map[string]string{"error": "不支持的文件格式"})
		return
	}
	if file.Size > maxCoverSize {
		c.JSON(consts.StatusBadRequest, map[string]string{"error": "封面文件不能超过10MB"})
	}

	// 4. 文件夹分类
	dateDir := time.Now().Format("20060102")
	// 物理路径： .run/uploads/covers/12345/20260715
	saveDir := filepath.Join(".run", "uploads", "covers", fmt.Sprintf("%d", authorID), dateDir)
	if err := os.MkdirAll(saveDir, 0755); err != nil {
		c.JSON(
			consts.StatusInternalServerError, map[string]string{"error": "创建上传目录失败"},
		)
		return
	}

	fileName := fmt.Sprintf("%d%s", time.Now().UnixNano(), ext) // 1784123456789123.mp4
	savePath := filepath.Join(saveDir, fileName)

	// 5. 将文件流写入硬盘
	if err := c.SaveUploadedFile(file, savePath); err != nil {
		c.JSON(consts.StatusInternalServerError, map[string]string{"error": "文件写入硬盘失败"})
		return
	}

	// 6. 返回 URL 给前端
	cover_url := fmt.Sprintf("http://127.0.0.1:20000/static/covers/%d/%s/%s", authorID, dateDir, fileName)
	c.JSON(consts.StatusOK, map[string]string{
		"message":   "上传成功",
		"cover_url": cover_url,
	})
}
func UploadVideo(ctx context.Context, c *app.RequestContext) {
	// 1. 读取 accountID
	accountIDVal, exists := c.Get("accountID")
	if !exists {
		c.JSON(consts.StatusUnauthorized, map[string]string{"error": "未授权"})
		return
	}
	authorID := accountIDVal.(int64)

	// 2. 接收前端的 "file" 文件
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(consts.StatusBadRequest, map[string]string{"error": "请上传 file 字段'"})
		return
	}

	// 3. 校验后缀名和大小
	ext := strings.ToLower(filepath.Ext(file.Filename))
	// 根据后缀自动归档
	if ext != ".mp4" {
		c.JSON(consts.StatusBadRequest, map[string]string{"error": "不支持的文件格式"})
		return
	}
	if file.Size > maxVideoSize {
		c.JSON(consts.StatusBadRequest, map[string]string{"error": "视频文件不能超过200MB"})
	}

	// 4. 文件夹分类
	dateDir := time.Now().Format("20060102")
	// 物理路径： .run/uploads/videos/12345/20260715
	saveDir := filepath.Join(".run", "uploads", "videos", fmt.Sprintf("%d", authorID), dateDir)
	if err := os.MkdirAll(saveDir, 0755); err != nil {
		c.JSON(consts.StatusInternalServerError, map[string]string{"error": "创建上传目录失败"})
		return
	}

	fileName := fmt.Sprintf("%d%s", time.Now().UnixNano(), ext) // 1784123456789123.mp4
	savePath := filepath.Join(saveDir, fileName)

	// 5. 将文件流写入硬盘
	if err := c.SaveUploadedFile(file, savePath); err != nil {
		c.JSON(consts.StatusInternalServerError, map[string]string{"error": "文件写入硬盘失败"})
		return
	}

	// 6. 返回 URL 给前端
	video_url := fmt.Sprintf("http://127.0.0.1:20000/static/videos/%d/%s/%s", authorID, dateDir, fileName)
	c.JSON(consts.StatusOK, map[string]string{
		"message":   "视频上传成功",
		"video_url": video_url,
	})
}
