package service

import (
	"context"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"rag-online-course/internal/repository/postgres"
	minioSvc "rag-online-course/internal/service/minio"
)

// ResourceService 负责资源对象业务编排。
type ResourceService struct {
	repo   *postgres.ResourceRepository
	minio  *minioSvc.Service
	urlTTL time.Duration
}

// NewResourceService 创建资源服务，统一封装资源入库、预签名上传与预览链接生成策略。
func NewResourceService(repo *postgres.ResourceRepository, minio *minioSvc.Service) *ResourceService {
	return &ResourceService{repo: repo, minio: minio, urlTTL: 15 * time.Minute}
}

// InitUpload 初始化上传前校验教师、课程、章节归属关系。
func (s *ResourceService) InitUpload(ctx context.Context, courseID, chapterID, teacherID int64, fileName string) (objectKey, uploadURL string, expireSeconds int, err error) {
	if err = s.repo.ValidateUploadScope(ctx, chapterID, courseID, teacherID); err != nil {
		return "", "", 0, err
	}
	objectKey = s.minio.NewObjectKey(strconv.FormatInt(courseID, 10), strconv.FormatInt(chapterID, 10), filepath.Base(fileName))
	presignedPutURL, err := s.minio.PresignedPutURL(ctx, objectKey, s.urlTTL)
	if err != nil {
		return "", "", 0, err
	}
	return objectKey, presignedPutURL.String(), int(s.urlTTL.Seconds()), nil
}

// ConfirmResource 在服务层根据 object_key 生成持久化 object_url，避免前端拼接不一致。
func (s *ResourceService) ConfirmResource(ctx context.Context, chapterID, teacherID int64, title, resourceType, objectKey, mimeType string, sizeBytes int64, sortOrder int) (string, error) {
	objectURL := s.minio.ObjectURL(objectKey)
	newResourceID, err := s.repo.ConfirmResource(ctx, chapterID, teacherID, title, resourceType, objectKey, objectURL, mimeType, sizeBytes, sortOrder)
	if err != nil {
		return "", err
	}
	return strconv.FormatInt(newResourceID, 10), nil
}

func (s *ResourceService) ReorderResource(ctx context.Context, resourceID, teacherID int64, sortOrder int) error {
	return s.repo.ReorderResource(ctx, resourceID, teacherID, sortOrder)
}

// ListResources 查询教师可见的章节资源列表。
func (s *ResourceService) ListResources(ctx context.Context, chapterID, teacherID int64) ([]map[string]any, error) {
	items, err := s.repo.ListResourcesByChapter(ctx, chapterID, teacherID)
	if err != nil {
		return nil, err
	}
	for _, item := range items {
		objectKey := toString(item["object_key"])
		if objectKey == "" {
			continue
		}
		contentType := s.previewContentType(
			toString(item["resource_type"]),
			toString(item["mime_type"]),
			objectKey,
		)
		previewURL, previewErr := s.minio.PresignedPreviewURL(ctx, objectKey, contentType, s.urlTTL)
		if previewErr != nil {
			// 预览链接生成失败时不阻断列表查询，前端可退回 object_url。
			continue
		}
		item["preview_url"] = previewURL.String()
	}
	return items, nil
}

// GetResourcePreviewURL 为资源生成“预览用” URL。
// 对 Office（doc/docx/ppt/pptx）会先将原文件转换为 PDF（仅用于预览），从而尽量避免内嵌时直接下载/渲染失败。
// 原始文件下载仍由调用方使用 object_url 完成。
func (s *ResourceService) GetResourcePreviewURL(ctx context.Context, resourceID, teacherID int64) (string, error) {
	item, err := s.repo.GetResourceByID(ctx, resourceID, teacherID)
	if err != nil {
		return "", err
	}

	objectKey := toString(item["object_key"])
	if objectKey == "" {
		return "", postgres.ErrNotFound
	}

	resourceType := toString(item["resource_type"])
	mimeType := toString(item["mime_type"])

	// Office -> PDF（预览用途）
	if resourceType == "doc" || resourceType == "docx" || resourceType == "ppt" {
		previewKey, convErr := s.minio.EnsureOfficePreviewPDF(ctx, objectKey)
		if convErr != nil {
			return "", fmt.Errorf("office preview convert failed: %w", convErr)
		}
		u, err := s.minio.PresignedPreviewURL(ctx, previewKey, "application/pdf", s.urlTTL)
		if err != nil {
			return "", err
		}
		return u.String(), nil
	}

	contentType := s.previewContentType(resourceType, mimeType, objectKey)
	previewURL, err := s.minio.PresignedPreviewURL(ctx, objectKey, contentType, s.urlTTL)
	if err != nil {
		return "", err
	}
	return previewURL.String(), nil
}

// DeleteResource 删除资源。
func (s *ResourceService) DeleteResource(ctx context.Context, resourceID, teacherID int64) error {
	return s.repo.DeleteResource(ctx, resourceID, teacherID)
}

// previewContentType 按资源类型与扩展名推断预览响应头，尽量避免浏览器走下载分支。
func (s *ResourceService) previewContentType(resourceType, mimeType, objectKey string) string {
	normalized := strings.TrimSpace(strings.ToLower(mimeType))
	if strings.HasPrefix(normalized, "text/") {
		if strings.Contains(normalized, "charset=") {
			return normalized
		}
		return normalized + "; charset=utf-8"
	}
	if normalized != "" && normalized != "application/octet-stream" {
		return normalized
	}
	lowerKey := strings.ToLower(objectKey)
	if strings.HasSuffix(lowerKey, ".md") || strings.HasSuffix(lowerKey, ".markdown") {
		return "text/markdown; charset=utf-8"
	}
	if strings.HasSuffix(lowerKey, ".txt") || resourceType == "txt" {
		return "text/plain; charset=utf-8"
	}
	if strings.HasSuffix(lowerKey, ".pdf") || resourceType == "pdf" {
		return "application/pdf"
	}
	if strings.HasSuffix(lowerKey, ".doc") || resourceType == "doc" {
		return "application/msword"
	}
	if strings.HasSuffix(lowerKey, ".docx") || resourceType == "docx" {
		return "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
	}
	if strings.HasSuffix(lowerKey, ".ppt") {
		return "application/vnd.ms-powerpoint"
	}
	if strings.HasSuffix(lowerKey, ".pptx") || resourceType == "ppt" {
		return "application/vnd.openxmlformats-officedocument.presentationml.presentation"
	}
	if strings.HasSuffix(lowerKey, ".mp3") || resourceType == "audio" {
		// 默认用 audio/mpeg；如 mime_type 已为其它 audio/*，上方将直接返回 mime_type。
		return "audio/mpeg"
	}
	if strings.HasSuffix(lowerKey, ".wav") {
		return "audio/wav"
	}
	if strings.HasSuffix(lowerKey, ".ogg") || strings.HasSuffix(lowerKey, ".oga") {
		return "audio/ogg"
	}
	if strings.HasSuffix(lowerKey, ".m4a") {
		return "audio/mp4"
	}
	if strings.HasSuffix(lowerKey, ".aac") {
		return "audio/aac"
	}
	if strings.HasSuffix(lowerKey, ".flac") {
		return "audio/flac"
	}
	if resourceType == "video" {
		return "video/mp4"
	}
	return "application/octet-stream"
}

func toString(v any) string {
	switch t := v.(type) {
	case string:
		return t
	case []byte:
		return string(t)
	default:
		return ""
	}
}
