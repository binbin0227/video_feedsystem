package handler

import (
	"encoding/binary"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"video_feedsystem/pkg/apperr"

	"github.com/cloudwego/hertz/pkg/app"
)

// detectUploadedContentType 根据文件开头的实际字节判断媒体类型，不能只相信用户提供的扩展名。
func detectUploadedContentType(file *multipart.FileHeader) (string, error) {
	source, err := file.Open()
	if err != nil {
		return "", apperr.Wrap(apperr.KindInvalid, "读取上传文件失败", err)
	}
	defer source.Close()

	header := make([]byte, 512)
	readCount, err := source.Read(header)
	if err != nil && err != io.EOF {
		return "", apperr.Wrap(apperr.KindInvalid, "读取上传文件失败", err)
	}
	if readCount == 0 {
		return "", apperr.New(apperr.KindInvalid, "上传文件不能为空")
	}
	// net/http 对 MP4 没有内置嗅探规则，因此额外识别 ISO Base Media 的 ftyp 文件头。
	if readCount >= 12 && string(header[4:8]) == "ftyp" {
		boxSize := binary.BigEndian.Uint32(header[:4])
		if boxSize == 1 || boxSize >= 12 {
			return "video/mp4", nil
		}
	}

	return http.DetectContentType(header[:readCount]), nil
}

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
