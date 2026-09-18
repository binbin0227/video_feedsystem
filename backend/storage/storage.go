package storage

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

var root string

// Init 初始化上传文件根目录
func Init(uploadRoot string) error {
	uploadRoot = strings.TrimSpace(uploadRoot)
	if uploadRoot == "" {
		return errors.New("上传文件根目录不能为空")
	}

	absoluteRoot, err := filepath.Abs(uploadRoot)
	if err != nil {
		return fmt.Errorf("解析上传文件根目录失败: %w", err)
	}

	if err := os.MkdirAll(absoluteRoot, 0755); err != nil {
		return fmt.Errorf("创建上传文件根目录失败: %w", err)
	}

	root = filepath.Clean(absoluteRoot)
	return nil
}

// Root 返回初始化后的上传文件根目录
func Root() string {
	return root
}

// ResolveUploadURL 把 /uploads/... 地址转换成本地文件路径
func ResolveUploadURL(uploadURL string) string {
	relativePath := strings.TrimPrefix(uploadURL, "/")
	return filepath.Join(root, filepath.FromSlash(relativePath))
}
