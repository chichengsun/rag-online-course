package app

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"go.uber.org/dig"
	"rag-online-course/internal/api"
	"rag-online-course/internal/api/handlers"
	"rag-online-course/internal/config"
	"rag-online-course/internal/integration/docreaderhttp"
	"rag-online-course/internal/repository/postgres"
	"rag-online-course/internal/service"
	authSvc "rag-online-course/internal/service/auth"
	minioSvc "rag-online-course/internal/service/minio"
)

// BuildContainer 负责装配应用依赖并返回可直接使用的 DI 容器。
func BuildContainer(ctx context.Context) *dig.Container {
	container := dig.New()

	mustProvide := func(name string, constructor any) {
		if err := container.Provide(constructor); err != nil {
			panic(fmt.Errorf("provide %s: %w", name, err))
		}
	}

	mustProvide("ctx", func() context.Context { return ctx })
	mustProvide("config.Load", config.Load)
	mustProvide("postgres.NewDB", postgres.NewDB)
	mustProvide("postgres.UserRepository", postgres.NewUserRepository)
	mustProvide("postgres.CourseRepository", postgres.NewCourseRepository)
	mustProvide("postgres.ChapterRepository", postgres.NewChapterRepository)
	mustProvide("postgres.ResourceRepository", postgres.NewResourceRepository)
	mustProvide("postgres.EnrollmentRepository", postgres.NewEnrollmentRepository)
	mustProvide("postgres.CatalogRepository", postgres.NewCatalogRepository)
	mustProvide("postgres.ProgressRepository", postgres.NewProgressRepository)
	mustProvide("postgres.EmbeddingChunkRepository", postgres.NewEmbeddingChunkRepository)
	mustProvide("postgres.TeacherAIModelRepository", postgres.NewTeacherAIModelRepository)
	mustProvide("postgres.KnowledgeChatRepository", postgres.NewKnowledgeChatRepository)
	mustProvide("postgres.KnowledgeRetrievalRepository", postgres.NewKnowledgeRetrievalRepository)
	mustProvide("http.ClientEmbedding", func() *http.Client {
		return &http.Client{Timeout: 120 * time.Second}
	})
	mustProvide("jwt.JWTService", authSvc.NewJWTService)
	mustProvide("auth.SessionStore", authSvc.NewSessionStore)
	mustProvide("minio.Service", minioSvc.New)
	mustProvide("docreaderhttp.Client", func(cfg config.Config) *docreaderhttp.Client {
		sec := cfg.DocReader.TimeoutSeconds
		if sec <= 0 {
			sec = 600
		}
		return docreaderhttp.NewClient(cfg.DocReader.BaseURL, cfg.DocReader.InternalToken, time.Duration(sec)*time.Second)
	})
	mustProvide("service.UserService", service.NewUserService)
	mustProvide("service.CourseService", service.NewCourseService)
	mustProvide("service.ChapterService", service.NewChapterService)
	mustProvide("service.ResourceService", service.NewResourceService)
	mustProvide("service.ResourceParseService", service.NewResourceParseService)
	mustProvide("service.EnrollmentService", service.NewEnrollmentService)
	mustProvide("service.CatalogService", service.NewCatalogService)
	mustProvide("service.ProgressService", service.NewProgressService)
	mustProvide("service.KnowledgeService", service.NewKnowledgeService)
	mustProvide("service.AIModelService", service.NewAIModelService)
	mustProvide("service.KnowledgeChatService", service.NewKnowledgeChatService)
	mustProvide("handlers.NewUserHandler", handlers.NewUserHandler)
	mustProvide("handlers.NewCourseHandler", handlers.NewCourseHandler)
	mustProvide("handlers.NewChapterHandler", handlers.NewChapterHandler)
	mustProvide("handlers.NewResourceHandler", handlers.NewResourceHandler)
	mustProvide("handlers.NewEnrollmentHandler", handlers.NewEnrollmentHandler)
	mustProvide("handlers.NewCatalogHandler", handlers.NewCatalogHandler)
	mustProvide("handlers.NewProgressHandler", handlers.NewProgressHandler)
	mustProvide("handlers.NewKnowledgeHandler", handlers.NewKnowledgeHandler)
	mustProvide("handlers.NewAIModelHandler", handlers.NewAIModelHandler)
	mustProvide("handlers.NewKnowledgeChatHandler", handlers.NewKnowledgeChatHandler)
	mustProvide("api.NewRouter", api.NewRouter)
	mustProvide("api.NewHTTPServer", api.NewHTTPServer)
	return container
}
