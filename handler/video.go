package handler

import (
	"context"
	"path/filepath"
	"strconv"
	"strings"

	"video_feedsystem/pkg/apperr"
	"video_feedsystem/pkg/httpx"
	"video_feedsystem/service"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
)

const (
	maxVideoSize int64 = 200 << 20
	maxCoverSize int64 = 10 << 20
)

type PublishRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	PlayURL     string `json:"play_url"`
	CoverURL    string `json:"cover_url"`
}

type AuthorVideoListResponse struct {
	Videos     []VideoResponse `json:"videos"`
	NextCursor string          `json:"next_cursor"`
	HasMore    bool            `json:"has_more"`
}

// PublishVideo 将已经上传的视频和封面信息写入数据库。
func PublishVideo(ctx context.Context, c *app.RequestContext) {
	// 1. 解析 JSON
	var req PublishRequest
	if err := c.BindAndValidate(&req); err != nil {
		httpx.WriteError(ctx, c, apperr.New(apperr.KindInvalid, "JSON 解析失败"))
		return
	}

	// 2. 从 c 中读取 accountID
	authorID, err := getAccountID(c)
	if err != nil {
		httpx.WriteError(ctx, c, err)
		return
	}

	// service.PublishVideo
	video, err := service.PublishVideo(ctx, authorID, req.Title, req.Description, req.PlayURL, req.CoverURL)
	if err != nil {
		httpx.WriteError(ctx, c, err)
		return
	}

	c.JSON(consts.StatusOK, map[string]any{
		"message": "视频发布成功！",
		"video":   newVideoResponse(video),
	})
}

// UploadCover 上传 jpg、jpeg 或 png 封面。
func UploadCover(ctx context.Context, c *app.RequestContext) {
	// 1. 从 c 中读取 accountID
	authorID, err := getAccountID(c)
	if err != nil {
		httpx.WriteError(ctx, c, err)
		return
	}

	// 2. 接收前端的 "file" 文件
	file, err := c.FormFile("file")
	if err != nil {
		httpx.WriteError(ctx, c, apperr.New(apperr.KindInvalid, "请上传 file 字段"))
		return
	}

	// 3. 校验后缀名和大小
	ext := strings.ToLower(filepath.Ext(file.Filename))
	if ext != ".jpg" && ext != ".jpeg" && ext != ".png" {
		httpx.WriteError(ctx, c, apperr.New(apperr.KindInvalid, "封面只支持 jpg、jpeg、png 格式"))
		return
	}
	if file.Size <= 0 {
		httpx.WriteError(ctx, c, apperr.New(apperr.KindInvalid, "封面文件不能为空"))
		return
	}
	if file.Size > maxCoverSize {
		httpx.WriteError(ctx, c, apperr.New(apperr.KindInvalid, "封面文件不能超过10MB"))
		return
	}

	// 3. 保存文件
	coverURL, err := saveUploadedFile(c, file, authorID, "covers", ext)
	if err != nil {
		httpx.WriteError(ctx, c, err)
		return
	}

	// 4. 返回 URL
	c.JSON(consts.StatusOK, map[string]string{
		"message":   "封面上传成功",
		"cover_url": coverURL,
	})
}

// UploadVideo 上传 mp4 视频。
func UploadVideo(ctx context.Context, c *app.RequestContext) {
	// 1. 从 c 中读取 accountID
	authorID, err := getAccountID(c)
	if err != nil {
		httpx.WriteError(ctx, c, err)
		return
	}

	// 2. 接收前端的 "file" 文件
	file, err := c.FormFile("file")
	if err != nil {
		httpx.WriteError(ctx, c, apperr.New(apperr.KindInvalid, "请上传 file 字段"))
		return
	}

	// 3. 校验后缀名和大小
	ext := strings.ToLower(filepath.Ext(file.Filename))
	if ext != ".mp4" {
		httpx.WriteError(ctx, c, apperr.New(apperr.KindInvalid, "视频只支持 mp4 格式"))
		return
	}
	if file.Size <= 0 {
		httpx.WriteError(ctx, c, apperr.New(apperr.KindInvalid, "视频文件不能为空"))
		return
	}
	if file.Size > maxVideoSize {
		httpx.WriteError(ctx, c, apperr.New(apperr.KindInvalid, "视频文件不能超过200MB"))
		return
	}

	// 3. 保存文件
	videoURL, err := saveUploadedFile(c, file, authorID, "videos", ext)
	if err != nil {
		httpx.WriteError(ctx, c, err)
		return
	}

	// 4. 返回 URL
	c.JSON(consts.StatusOK, map[string]string{
		"message":   "视频上传成功",
		"video_url": videoURL,
	})
}

// ListByAuthorID 分页查询指定作者发布的视频。
func ListByAuthorID(ctx context.Context, c *app.RequestContext) {
	// 1. 解析 author_id
	authorID, err := parsePositiveInt64Query(c, "author_id")
	if err != nil {
		httpx.WriteError(ctx, c, err)
		return
	}

	// 2. 解析 cursor 和 limit
	cursor, err := parseOptionalCursor(c)
	if err != nil {
		httpx.WriteError(ctx, c, err)
		return
	}

	limit, err := parseOptionalLimit(c)
	if err != nil {
		httpx.WriteError(ctx, c, err)
		return
	}

	// 3. 查询作者发布的视频
	result, err := service.ListByAuthorID(ctx, authorID, cursor, limit)
	if err != nil {
		httpx.WriteError(ctx, c, err)
		return
	}

	// 4. 将下一页游标转换为字符串
	nextCursor := ""
	if result.NextCursor > 0 {
		nextCursor = strconv.FormatInt(result.NextCursor, 10)
	}

	// 5. 返回结果
	c.JSON(consts.StatusOK, AuthorVideoListResponse{
		Videos:     newVideoListResponse(result.Videos),
		NextCursor: nextCursor,
		HasMore:    result.HasMore,
	})
}

// GetVideoDetail 查询单个视频详情。
func GetVideoDetail(ctx context.Context, c *app.RequestContext) {
	// 1. 从 c 中读取 video_id 并转换为 int64
	videoID, err := parsePositiveInt64Query(c, "video_id")
	if err != nil {
		httpx.WriteError(ctx, c, err)
		return
	}

	// service.GetVideoDetail
	video, err := service.GetVideoDetail(ctx, videoID)
	if err != nil {
		httpx.WriteError(ctx, c, err)
		return
	}

	c.JSON(consts.StatusOK, map[string]any{"video": newVideoResponse(video)})
}
