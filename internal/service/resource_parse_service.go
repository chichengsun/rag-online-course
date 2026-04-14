package service

import (
	"context"
	"errors"
	"fmt"

	"rag-online-course/internal/config"
	"rag-online-course/internal/dto/course"
	"rag-online-course/internal/integration/docreaderhttp"
	"rag-online-course/internal/repository/postgres"

	"github.com/sirupsen/logrus"
)

// ResourceParseService 编排教师触发的资源解析：校验归属、调用 docreader-http；结果不落库（分块/向量见 resource_embedding_chunks 管线）。
type ResourceParseService struct {
	resRepo *postgres.ResourceRepository
	client  *docreaderhttp.Client
	bucket  string
	// useOCR 是否在调用 docreader 时携带 use_ocr（由配置 docreader.use_ocr 决定；网关需启用 OCR_BACKEND）。
	useOCR bool
}

// NewResourceParseService 构造解析编排服务；bucket 与 MinIO 课程资源桶一致。
func NewResourceParseService(
	resRepo *postgres.ResourceRepository,
	client *docreaderhttp.Client,
	cfg config.Config,
) *ResourceParseService {
	return &ResourceParseService{
		resRepo: resRepo,
		client:  client,
		bucket:  cfg.Minio.Bucket,
		useOCR:  cfg.DocReader.UseOCR,
	}
}

// ParseResource 对指定资源触发解析并返回网关结果；持久化由后续分块/嵌入接口写入 resource_embedding_chunks。
func (s *ResourceParseService) ParseResource(ctx context.Context, resourceID, teacherID int64) (*course.ParseResourceResp, error) {
	item, err := s.resRepo.GetResourceByID(ctx, resourceID, teacherID)
	if err != nil {
		return nil, err
	}
	objectKey := toString(item["object_key"])
	if objectKey == "" {
		return nil, postgres.ErrNotFound
	}

	readResp, err := s.client.V1Read(ctx, docreaderhttp.V1ReadRequest{
		ObjectKey: objectKey,
		Bucket:    s.bucket,
		UseOCR:    s.useOCR,
	})
	if errors.Is(err, docreaderhttp.ErrNotConfigured) {
		return nil, fmt.Errorf("docreader 未配置：请在配置中设置 docreader.base_url")
	}
	if err != nil {
		logrus.WithError(err).WithField("resource_id", resourceID).Warn("docreader request failed")
		return nil, err
	}

	if readResp.Error != "" {
		return &course.ParseResourceResp{
			Status:     "failed",
			Markdown:   readResp.Markdown,
			Metadata:   readResp.Metadata,
			Error:      readResp.Error,
			ImageCount: len(readResp.Images),
		}, nil
	}

	return &course.ParseResourceResp{
		Status:     "ok",
		Markdown:   readResp.Markdown,
		Metadata:   readResp.Metadata,
		ImageCount: len(readResp.Images),
	}, nil
}
