package handler

import (
	"fmt"
	"mime/multipart"
	"os"
	"path/filepath"
	"time"

	"video_feedsystem/pkg/apperr"
	"video_feedsystem/storage"

	"github.com/cloudwego/hertz/pkg/app"
)

// 保存文件并返回相对访问路径
func saveUploadedFile(c *app.RequestContext, file *multipart.FileHeader, authorID int64, category, ext string) (string, error) {
	dateDir := time.Now().Format("20060102")
	saveDir := filepath.Join(storage.Root(), "uploads", category, fmt.Sprintf("%d", authorID), dateDir)
	if err := os.MkdirAll(saveDir, 0755); err != nil {
		return "", apperr.Wrap(apperr.KindInternal, "创建上传目录失败", err)
	}
	fileName := fmt.Sprintf("%d%s", time.Now().UnixNano(), ext)
	savePath := filepath.Join(saveDir, fileName)
	if err := c.SaveUploadedFile(file, savePath); err != nil {
		return "", apperr.Wrap(apperr.KindInternal, "文件保存失败", err)
	}

	return fmt.Sprintf("/uploads/%s/%d/%s/%s", category, authorID, dateDir, fileName), nil
}
