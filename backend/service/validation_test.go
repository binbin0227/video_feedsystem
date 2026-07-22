package service

import (
	"context"
	"errors"
	"strings"
	"testing"

	"video_feedsystem/pkg/apperr"
)

func requireInvalidError(t *testing.T, err error) {
	t.Helper()

	var appError *apperr.AppError
	if !errors.As(err, &appError) {
		t.Fatalf("expected AppError, got %T", err)
	}
	if appError.Kind != apperr.KindInvalid {
		t.Fatalf("got error kind %q, want %q", appError.Kind, apperr.KindInvalid)
	}
}

func TestValidateUploadPath(t *testing.T) {
	tests := []struct {
		name       string
		uploadPath string
		category   string
		authorID   int64
		wantErr    bool
	}{
		{name: "valid video", uploadPath: "/uploads/videos/123/20260722/video.mp4", category: "videos", authorID: 123},
		{name: "wrong account", uploadPath: "/uploads/videos/456/20260722/video.mp4", category: "videos", authorID: 123, wantErr: true},
		{name: "wrong category", uploadPath: "/uploads/covers/123/20260722/video.mp4", category: "videos", authorID: 123, wantErr: true},
		{name: "path traversal", uploadPath: "/uploads/videos/123/../456/video.mp4", category: "videos", authorID: 123, wantErr: true},
		{name: "relative path", uploadPath: "uploads/videos/123/video.mp4", category: "videos", authorID: 123, wantErr: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := validateUploadPath(test.uploadPath, test.category, test.authorID)
			if test.wantErr {
				if err == nil {
					t.Fatal("expected an error")
				}
				requireInvalidError(t, err)
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestSearchAccountsRejectsInvalidKeywordBeforeDatabaseQuery(t *testing.T) {
	tests := []string{"   ", strings.Repeat("用", maxAccountSearchKeywordLength+1)}
	for _, keyword := range tests {
		_, err := SearchAccounts(context.Background(), keyword)
		if err == nil {
			t.Fatal("expected an error")
		}
		requireInvalidError(t, err)
	}
}
