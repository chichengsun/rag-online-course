package service

import (
	"context"

	dtoLearning "rag-online-course/internal/dto/learning"
	"rag-online-course/internal/repository/postgres"
)

// CatalogService 负责课程目录对象业务编排。
type CatalogService struct {
	repo *postgres.CatalogRepository
}

// NewCatalogService 创建目录业务服务。
func NewCatalogService(repo *postgres.CatalogRepository) *CatalogService {
	return &CatalogService{repo: repo}
}

// Catalog 获取课程目录（章节与资源）。
func (s *CatalogService) Catalog(ctx context.Context, courseID, studentID int64) (dtoLearning.CatalogResp, error) {
	catalogMap, err := s.repo.GetCatalog(ctx, courseID, studentID)
	if err != nil {
		return nil, err
	}
	return dtoLearning.CatalogResp(catalogMap), nil
}
