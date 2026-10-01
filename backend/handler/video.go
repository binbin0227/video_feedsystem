package handler

import (
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"video_feedsystem/pkg/apperr"
	"video_feedsystem/pkg/httpx"
	"video_feedsystem/service"

	"github.com/gin-gonic/gin"
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

// 将已经上传的视频和封面信息写入数据库
func PublishVideo(c *gin.Context) {
	var req PublishRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.WriteError(c, apperr.New(apperr.KindInvalid, "JSON 解析失败"))
		return
	}

	authorID, err := getAccountID(c)
	if err != nil {
		httpx.WriteError(c, err)
		return
	}

	video, err := service.PublishVideo(c.Request.Context(), authorID, req.Title, req.Description, req.PlayURL, req.CoverURL)
	if err != nil {
		httpx.WriteError(c, err)
		return
	}

	c.JSON(http.StatusOK, map[string]any{
		"message": "视频发布成功！",
		"video":   newVideoResponse(video),
	})
}

// 上传 jpg、jpeg 或 png 封面
func UploadCover(c *gin.Context) {
	authorID, err := getAccountID(c)
	if err != nil {
		httpx.WriteError(c, err)
		return
	}

	file, err := c.FormFile("file")
	if err != nil {
		httpx.WriteError(c, apperr.New(apperr.KindInvalid, "请上传 file 字段"))
		return
	}

	ext := strings.ToLower(filepath.Ext(file.Filename))
	if ext != ".jpg" && ext != ".jpeg" && ext != ".png" {
		httpx.WriteError(c, apperr.New(apperr.KindInvalid, "封面只支持 jpg、jpeg、png 格式"))
		return
	}
	if file.Size <= 0 {
		httpx.WriteError(c, apperr.New(apperr.KindInvalid, "封面文件不能为空"))
		return
	}
	if file.Size > maxCoverSize {
		httpx.WriteError(c, apperr.New(apperr.KindInvalid, "封面文件不能超过10MB"))
		return
	}

	coverURL, err := saveUploadedFile(c, file, authorID, "covers", ext)
	if err != nil {
		httpx.WriteError(c, err)
		return
	}
	c.JSON(http.StatusOK, map[string]string{
		"message":   "封面上传成功",
		"cover_url": coverURL,
	})
}

// 上传 mp4 视频
func UploadVideo(c *gin.Context) {
	authorID, err := getAccountID(c)
	if err != nil {
		httpx.WriteError(c, err)
		return
	}

	file, err := c.FormFile("file")
	if err != nil {
		httpx.WriteError(c, apperr.New(apperr.KindInvalid, "请上传 file 字段"))
		return
	}

	ext := strings.ToLower(filepath.Ext(file.Filename))
	if ext != ".mp4" {
		httpx.WriteError(c, apperr.New(apperr.KindInvalid, "视频只支持 mp4 格式"))
		return
	}
	if file.Size <= 0 {
		httpx.WriteError(c, apperr.New(apperr.KindInvalid, "视频文件不能为空"))
		return
	}
	if file.Size > maxVideoSize {
		httpx.WriteError(c, apperr.New(apperr.KindInvalid, "视频文件不能超过200MB"))
		return
	}

	videoURL, err := saveUploadedFile(c, file, authorID, "videos", ext)
	if err != nil {
		httpx.WriteError(c, err)
		return
	}
	c.JSON(http.StatusOK, map[string]string{
		"message":   "视频上传成功",
		"video_url": videoURL,
	})
}

// 分页查询指定作者发布的视频
func ListByAuthorID(c *gin.Context) {
	authorID, err := parsePositiveInt64Query(c, "author_id")
	if err != nil {
		httpx.WriteError(c, err)
		return
	}

	cursor, err := parseOptionalCursor(c)
	if err != nil {
		httpx.WriteError(c, err)
		return
	}

	limit, err := parseOptionalLimit(c)
	if err != nil {
		httpx.WriteError(c, err)
		return
	}

	result, err := service.ListByAuthorID(c.Request.Context(), authorID, cursor, limit)
	if err != nil {
		httpx.WriteError(c, err)
		return
	}
	nextCursor := ""
	if result.NextCursor > 0 {
		nextCursor = strconv.FormatInt(result.NextCursor, 10)
	}
	c.JSON(http.StatusOK, AuthorVideoListResponse{
		Videos:     newVideoListResponse(result.Videos),
		NextCursor: nextCursor,
		HasMore:    result.HasMore,
	})
}

// 查询单个视频详情
func GetVideoDetail(c *gin.Context) {
	videoID, err := parsePositiveInt64Query(c, "video_id")
	if err != nil {
		httpx.WriteError(c, err)
		return
	}

	video, err := service.GetVideoDetail(c.Request.Context(), videoID)
	if err != nil {
		httpx.WriteError(c, err)
		return
	}
	c.JSON(http.StatusOK, map[string]any{"video": newVideoResponse(video)})
}
