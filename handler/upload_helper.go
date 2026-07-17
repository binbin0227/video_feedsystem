package handler

import (
	"fmt"
	"mime/multipart"
	"os"
	"path/filepath"
	"time"

	"video_feedsystem/pkg/apperr"

	"github.com/cloudwego/hertz/pkg/app"
)

// saveUploadedFile 保存文件，并返回可存入数据库的相对访问路径。
func saveUploadedFile(c *app.RequestContext, file *multipart.FileHeader, authorID int64, category, ext string) (string, error) {
	// 1. 定义保存路径
	dateDir := time.Now().Format("20060102")
	saveDir := filepath.Join(".run", "uploads", category, fmt.Sprintf("%d", authorID), dateDir)

	// 2. 创建目录
	if err := os.MkdirAll(saveDir, 0755); err != nil {
		return "", apperr.Wrap(apperr.KindInternal, "创建上传目录失败", err)
	}

	// 3. 保存文件
	fileName := fmt.Sprintf("%d%s", time.Now().UnixNano(), ext)
	savePath := filepath.Join(saveDir, fileName)
	if err := c.SaveUploadedFile(file, savePath); err != nil {
		return "", apperr.Wrap(apperr.KindInternal, "文件保存失败", err)
	}

	return fmt.Sprintf("/uploads/%s/%d/%s/%s", category, authorID, dateDir, fileName), nil
}
