package api

import (
	"net/http"

	"rag-online-course/internal/api/handlers"
	"rag-online-course/internal/api/middleware"
	"rag-online-course/internal/config"
	"rag-online-course/internal/service/auth"

	"github.com/gin-gonic/gin"
)

// NewRouter 注册系统全部路由，按资源对象组织处理器并通过鉴权中间件控制访问边界。
// 其中 Redis Session 与 JWT 一起用于校验请求是否来自有效登录会话。
func NewRouter(userH *handlers.UserHandler, courseH *handlers.CourseHandler, chapterH *handlers.ChapterHandler, sectionH *handlers.SectionHandler, resourceH *handlers.ResourceHandler, enrollmentH *handlers.EnrollmentHandler, catalogH *handlers.CatalogHandler, progressH *handlers.ProgressHandler, knowledgeH *handlers.KnowledgeHandler, aiModelH *handlers.AIModelHandler, knowledgeChatH *handlers.KnowledgeChatHandler, courseDesignH *handlers.CourseDesignHandler, questionBankH *handlers.QuestionBankHandler, jwtSvc *auth.JWTService, sessionStore *auth.SessionStore) *gin.Engine {
	r := gin.New()
	r.Use(
		middleware.CORS(),
		middleware.RequestLogger(),
		middleware.Recovery(),
		middleware.ErrorHandler(),
	)

	v1 := r.Group("/api/v1")
	{
		authGroup := v1.Group("/auth")
		{
			authGroup.POST("/register", userH.Register)
			authGroup.POST("/login", userH.Login)
			authGroup.POST("/refresh", userH.Refresh)
		}

		protected := v1.Group("/")
		protected.Use(middleware.RequireAuth(jwtSvc, sessionStore))
		{
			protected.GET("/me", userH.Me)
		}

		// Course 相关接口（需 teacher 角色）。
		teacher := v1.Group("/teacher")
		teacher.Use(middleware.RequireAuth(jwtSvc, sessionStore), middleware.RequireRole("teacher"))
		{
			teacher.GET("/courses", courseH.ListCourses)
			teacher.POST("/courses", courseH.CreateCourse)
			teacher.PUT("/courses/:courseId", courseH.UpdateCourse)
			teacher.DELETE("/courses/:courseId", courseH.DeleteCourse)
			teacher.POST("/courses/:courseId/design/outline-draft/generate", courseDesignH.GenerateOutlineDraft)
			teacher.POST("/courses/:courseId/design/outline-draft/apply", courseDesignH.ApplyOutlineDraft)
			teacher.GET("/courses/:courseId/question-bank", questionBankH.ListByCourse)
			teacher.POST("/courses/:courseId/question-bank", questionBankH.Create)
			teacher.POST("/courses/:courseId/question-bank/import/parse", questionBankH.ParseImportFromFile)
			teacher.POST("/courses/:courseId/question-bank/import/confirm", questionBankH.ConfirmImportBatch)
			teacher.PUT("/question-bank/items/:itemId", questionBankH.Update)
			teacher.DELETE("/question-bank/items/:itemId", questionBankH.Delete)
			teacher.GET("/courses/:courseId/chapters", chapterH.ListChapters)
			teacher.POST("/courses/:courseId/chapters", chapterH.CreateChapter)
			teacher.GET("/courses/:courseId/chapters/:chapterId/sections", sectionH.ListSections)
			teacher.POST("/courses/:courseId/chapters/:chapterId/sections", sectionH.CreateSection)
			teacher.DELETE("/chapters/:chapterId", chapterH.DeleteChapter)
			teacher.PUT("/chapters/:chapterId", chapterH.UpdateChapter)
			teacher.PUT("/chapters/:chapterId/reorder", chapterH.ReorderChapter)
			teacher.DELETE("/sections/:sectionId", sectionH.DeleteSection)
			teacher.PUT("/sections/:sectionId", sectionH.UpdateSection)
			teacher.PUT("/sections/:sectionId/reorder", sectionH.ReorderSection)
			teacher.POST("/sections/:sectionId/lesson-plan/generate", sectionH.GenerateLessonPlanDraft)
			teacher.GET("/sections/:sectionId/resources", resourceH.ListResources)
			teacher.POST("/sections/:sectionId/resources/init-upload", resourceH.InitUpload)
			teacher.POST("/sections/:sectionId/resources/confirm", resourceH.ConfirmResource)
			teacher.GET("/resources/:resourceId", resourceH.GetResourceDetail)
			teacher.PUT("/resources/:resourceId", resourceH.UpdateResource)
			teacher.POST("/resources/:resourceId/parse", resourceH.ParseResource)
			teacher.POST("/resources/:resourceId/summarize", resourceH.SummarizeResource)
			teacher.GET("/courses/:courseId/knowledge/resources", knowledgeH.ListKnowledgeResources)
			teacher.POST("/resources/:resourceId/knowledge/chunk-preview", knowledgeH.ChunkPreview)
			teacher.DELETE("/resources/:resourceId/knowledge/chunks/:chunkId", knowledgeH.DeleteKnowledgeChunk)
			teacher.PATCH("/resources/:resourceId/knowledge/chunks/:chunkId", knowledgeH.UpdateKnowledgeChunk)
			teacher.DELETE("/resources/:resourceId/knowledge/chunks", knowledgeH.ClearKnowledgeChunks)
			teacher.PUT("/resources/:resourceId/knowledge/chunks", knowledgeH.SaveKnowledgeChunks)
			teacher.POST("/resources/:resourceId/knowledge/chunks/confirm", knowledgeH.ConfirmKnowledgeChunks)
			teacher.GET("/resources/:resourceId/knowledge/chunks", knowledgeH.ListKnowledgeChunks)
			teacher.POST("/resources/:resourceId/knowledge/embed", knowledgeH.EmbedResource)
			teacher.POST("/courses/:courseId/knowledge/chats/sessions", knowledgeChatH.CreateSession)
			teacher.GET("/knowledge/chats/sessions", knowledgeChatH.ListSessions)
			teacher.PATCH("/knowledge/chats/sessions/:sessionId", knowledgeChatH.UpdateSession)
			teacher.DELETE("/knowledge/chats/sessions/:sessionId", knowledgeChatH.DeleteSession)
			teacher.GET("/knowledge/chats/sessions/:sessionId/messages", knowledgeChatH.ListMessages)
			teacher.POST("/knowledge/chats/sessions/:sessionId/ask", knowledgeChatH.AskInSession)
			teacher.POST("/knowledge/chats/sessions/:sessionId/ask/stream", knowledgeChatH.AskInSessionStream)
			teacher.GET("/ai-models", aiModelH.ListAIModels)
			teacher.POST("/ai-models/test-connection", aiModelH.TestAIModelConnection)
			teacher.POST("/ai-models", aiModelH.CreateAIModel)
			teacher.PUT("/ai-models/:modelId", aiModelH.UpdateAIModel)
			teacher.DELETE("/ai-models/:modelId", aiModelH.DeleteAIModel)
			teacher.PUT("/resources/:resourceId/reorder", resourceH.ReorderResource)
			teacher.DELETE("/resources/:resourceId", resourceH.DeleteResource)
		}

		// Enrollment/Catalog/Progress 相关接口（需 student 角色）。
		student := v1.Group("/")
		student.Use(middleware.RequireAuth(jwtSvc, sessionStore), middleware.RequireRole("student"))
		{
			student.GET("/courses", enrollmentH.ListCourses)
			student.POST("/courses/:courseId/enroll", enrollmentH.Enroll)
			student.GET("/my/courses", enrollmentH.MyCourses)
			student.GET("/my/courses/:courseId/catalog", catalogH.Catalog)
			student.POST("/my/resources/:resourceId/progress", progressH.UpdateProgress)
			student.POST("/my/resources/:resourceId/complete", progressH.CompleteResource)
		}
	}
	return r
}

// NewHTTPServer 根据配置与路由创建 HTTP Server。
func NewHTTPServer(cfg config.Config, router *gin.Engine) *http.Server {
	return &http.Server{
		Addr:    ":" + cfg.Server.Port,
		Handler: router,
	}
}
