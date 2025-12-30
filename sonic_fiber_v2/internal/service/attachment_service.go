package service

import (
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type AttachmentServiceImpl struct {
	mu          sync.RWMutex
	attachments []map[string]interface{}
	idSeq       int64
	uploadDir   string
}

func NewAttachmentService(uploadDir string) AttachmentService {
	return &AttachmentServiceImpl{
		attachments: make([]map[string]interface{}, 0),
		idSeq:       1,
		uploadDir:   uploadDir,
	}
}

func (s *AttachmentServiceImpl) Upload(ctx context.Context, filename string, src io.Reader, size int64) (map[string]interface{}, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 确保上传目录存在
	if err := os.MkdirAll(s.uploadDir, os.ModePerm); err != nil {
		return nil, err
	}

	// 生成唯一文件名
	ext := filepath.Ext(filename)
	timestamp := time.Now().Format("20060102150405")
	newFilename := timestamp + ext
	filepath := filepath.Join(s.uploadDir, newFilename)

	// 保存文件
	dst, err := os.Create(filepath)
	if err != nil {
		return nil, err
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		return nil, err
	}

	s.idSeq++
	attachment := map[string]interface{}{
		"id":         s.idSeq,
		"filename":   filename,
		"path":       "/uploads/" + newFilename,
		"size":       size,
		"created_at": time.Now(),
	}

	s.attachments = append(s.attachments, attachment)
	return attachment, nil
}

func (s *AttachmentServiceImpl) List(ctx context.Context, page, size int, keyword string) ([]map[string]interface{}, int64, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var filtered []map[string]interface{}
	for _, attachment := range s.attachments {
		if keyword != "" {
			filename, _ := attachment["filename"].(string)
			if len(filename) == 0 || len(filename) < len(keyword) || filename[:len(keyword)] != keyword {
				continue
			}
		}
		filtered = append(filtered, attachment)
	}

	total := int64(len(filtered))
	start := (page - 1) * size
	if start >= len(filtered) {
		return []map[string]interface{}{}, total, nil
	}

	end := start + size
	if end > len(filtered) {
		end = len(filtered)
	}

	return filtered[start:end], total, nil
}

func (s *AttachmentServiceImpl) Delete(ctx context.Context, id int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	var attachment map[string]interface{}
	var index int
	for i, a := range s.attachments {
		if a["id"] == id {
			attachment = a
			index = i
			break
		}
	}

	if attachment == nil {
		return errors.New("attachment not found")
	}

	// 删除文件
	path, _ := attachment["path"].(string)
	if path != "" {
		filePath := filepath.Join(s.uploadDir, filepath.Base(path))
		os.Remove(filePath)
	}

	s.attachments = append(s.attachments[:index], s.attachments[index+1:]...)
	return nil
}
