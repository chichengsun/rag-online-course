package integration_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"rag-online-course/internal/api"
	"rag-online-course/internal/api/handlers"
	"rag-online-course/internal/config"
	"rag-online-course/internal/integration/docreaderhttp"
	"rag-online-course/internal/repository/postgres"
	"rag-online-course/internal/service"
	authSvc "rag-online-course/internal/service/auth"
	minioSvc "rag-online-course/internal/service/minio"

	"gorm.io/gorm"
)

func TestAuthAndCourseFlow(t *testing.T) {
	if os.Getenv("RUN_INTEGRATION_TESTS") != "1" {
		t.Skip("set RUN_INTEGRATION_TESTS=1 to run integration tests")
	}

	ctx := context.Background()
	cfg := config.Load()

	db, err := postgres.NewDB(ctx, cfg)
	if err != nil {
		t.Fatalf("connect postgres failed: %v", err)
	}
	defer postgres.CloseDB(db)

	if err = runMigration(ctx, db); err != nil {
		t.Fatalf("run migration failed: %v", err)
	}

	jwt := authSvc.NewJWTService(cfg)
	sessionStore := authSvc.NewSessionStore(cfg)
	defer sessionStore.Close()

	// 清空会话库，避免历史脏数据影响断言。
	if err = flushRedis(ctx, cfg); err != nil {
		t.Fatalf("flush redis failed: %v", err)
	}

	minioClient, err := minioSvc.New(cfg)
	if err != nil {
		t.Fatalf("init minio service failed: %v", err)
	}

	userRepo := postgres.NewUserRepository(db)
	courseRepo := postgres.NewCourseRepository(db)
	chapterRepo := postgres.NewChapterRepository(db)
	resourceRepo := postgres.NewResourceRepository(db)
	sectionRepo := postgres.NewSectionRepository(db)
	docClient := docreaderhttp.NewClient("", "", time.Second)
	resourceParseSvc := service.NewResourceParseService(resourceRepo, docClient, cfg)
	enrollmentRepo := postgres.NewEnrollmentRepository(db)
	catalogRepo := postgres.NewCatalogRepository(db)
	progressRepo := postgres.NewProgressRepository(db)
	embeddingChunkRepo := postgres.NewEmbeddingChunkRepository(db)
	teacherAIModelRepo := postgres.NewTeacherAIModelRepository(db)
	questionBankRepo := postgres.NewQuestionBankRepository(db)
	chatRepo := postgres.NewKnowledgeChatRepository(db)
	retrievalRepo := postgres.NewKnowledgeRetrievalRepository(db)
	httpClient := &http.Client{Timeout: 30 * time.Second}
	userSvc := service.NewUserService(userRepo, jwt, sessionStore)
	courseSvc := service.NewCourseService(courseRepo)
	chapterSvc := service.NewChapterService(chapterRepo, sectionRepo)
	sectionSvc := service.NewSectionService(sectionRepo)
	resourceSvc := service.NewResourceService(resourceRepo, sectionRepo, minioClient)
	summarySvc := service.NewResourceAISummaryService(resourceParseSvc, resourceRepo, teacherAIModelRepo, httpClient)
	enrollmentSvc := service.NewEnrollmentService(courseRepo, enrollmentRepo)
	catalogSvc := service.NewCatalogService(catalogRepo)
	progressSvc := service.NewProgressService(progressRepo)
	knowledgeSvc := service.NewKnowledgeService(embeddingChunkRepo, resourceParseSvc, teacherAIModelRepo, httpClient)
	aiModelSvc := service.NewAIModelService(teacherAIModelRepo, httpClient)
	knowledgeChatSvc := service.NewKnowledgeChatService(chatRepo, retrievalRepo, teacherAIModelRepo, cfg, httpClient)
	courseDesignSvc := service.NewCourseDesignService(cfg, courseRepo, teacherAIModelRepo, chapterSvc, sectionSvc, httpClient)
	questionBankSvc := service.NewQuestionBankService(cfg, courseRepo, teacherAIModelRepo, questionBankRepo, httpClient)
	lessonPlanSvc := service.NewSectionLessonPlanService(cfg, sectionRepo, resourceRepo, teacherAIModelRepo, httpClient)

	userH := handlers.NewUserHandler(userSvc)
	courseH := handlers.NewCourseHandler(courseSvc)
	chapterH := handlers.NewChapterHandler(chapterSvc)
	sectionH := handlers.NewSectionHandler(sectionSvc, lessonPlanSvc)
	resourceH := handlers.NewResourceHandler(resourceSvc, resourceParseSvc, summarySvc)
	enrollmentH := handlers.NewEnrollmentHandler(enrollmentSvc)
	catalogH := handlers.NewCatalogHandler(catalogSvc)
	progressH := handlers.NewProgressHandler(progressSvc)
	knowledgeH := handlers.NewKnowledgeHandler(knowledgeSvc)
	aiModelH := handlers.NewAIModelHandler(aiModelSvc)
	knowledgeChatH := handlers.NewKnowledgeChatHandler(knowledgeChatSvc)
	courseDesignH := handlers.NewCourseDesignHandler(courseDesignSvc)
	questionBankH := handlers.NewQuestionBankHandler(questionBankSvc)
	router := api.NewRouter(userH, courseH, chapterH, sectionH, resourceH, enrollmentH, catalogH, progressH, knowledgeH, aiModelH, knowledgeChatH, courseDesignH, questionBankH, jwt, sessionStore)

	ts := httptest.NewServer(router)
	defer ts.Close()

	suffix := time.Now().UnixNano()
	registerBody := map[string]any{
		"email":    "teacher" + itoa(suffix) + "@example.com",
		"username": "teacher" + itoa(suffix),
		"name":     "测试教师",
		"password": "Passw0rd!",
		"role":     "teacher",
	}
	registerResp := requestJSON(t, http.MethodPost, ts.URL+"/api/v1/auth/register", "", registerBody)
	if registerResp.StatusCode != http.StatusCreated {
		t.Fatalf("register status=%d body=%s", registerResp.StatusCode, registerResp.BodyString)
	}

	loginBody := map[string]any{
		"account":  registerBody["username"],
		"password": registerBody["password"],
	}
	loginResp := requestJSON(t, http.MethodPost, ts.URL+"/api/v1/auth/login", "", loginBody)
	if loginResp.StatusCode != http.StatusOK {
		t.Fatalf("login status=%d body=%s", loginResp.StatusCode, loginResp.BodyString)
	}
	access := loginResp.JSON["access_token"].(string)
	if access == "" {
		t.Fatalf("empty access token, body=%s", loginResp.BodyString)
	}

	createCourseBody := map[string]any{
		"title":       " Go 入门 ",
		"description": "第一门课",
	}
	createCourseResp := requestJSON(t, http.MethodPost, ts.URL+"/api/v1/teacher/courses", access, createCourseBody)
	if createCourseResp.StatusCode != http.StatusCreated {
		t.Fatalf("create course status=%d body=%s", createCourseResp.StatusCode, createCourseResp.BodyString)
	}
	if createCourseResp.JSON["id"] == "" {
		t.Fatalf("create course id empty body=%s", createCourseResp.BodyString)
	}
}

func TestTeacherBuildAndStudentLearnFlow(t *testing.T) {
	if os.Getenv("RUN_INTEGRATION_TESTS") != "1" {
		t.Skip("set RUN_INTEGRATION_TESTS=1 to run integration tests")
	}

	ctx := context.Background()
	cfg := config.Load()

	db, err := postgres.NewDB(ctx, cfg)
	if err != nil {
		t.Fatalf("connect postgres failed: %v", err)
	}
	defer postgres.CloseDB(db)

	if err = runMigration(ctx, db); err != nil {
		t.Fatalf("run migration failed: %v", err)
	}
	if err = flushRedis(ctx, cfg); err != nil {
		t.Fatalf("flush redis failed: %v", err)
	}

	jwt := authSvc.NewJWTService(cfg)
	sessionStore := authSvc.NewSessionStore(cfg)
	defer sessionStore.Close()

	minioClient, err := minioSvc.New(cfg)
	if err != nil {
		t.Fatalf("init minio service failed: %v", err)
	}

	userRepo := postgres.NewUserRepository(db)
	courseRepo := postgres.NewCourseRepository(db)
	chapterRepo := postgres.NewChapterRepository(db)
	resourceRepo := postgres.NewResourceRepository(db)
	sectionRepo := postgres.NewSectionRepository(db)
	docClient := docreaderhttp.NewClient("", "", time.Second)
	resourceParseSvc := service.NewResourceParseService(resourceRepo, docClient, cfg)
	enrollmentRepo := postgres.NewEnrollmentRepository(db)
	catalogRepo := postgres.NewCatalogRepository(db)
	progressRepo := postgres.NewProgressRepository(db)
	embeddingChunkRepo := postgres.NewEmbeddingChunkRepository(db)
	teacherAIModelRepo := postgres.NewTeacherAIModelRepository(db)
	questionBankRepo := postgres.NewQuestionBankRepository(db)
	chatRepo := postgres.NewKnowledgeChatRepository(db)
	retrievalRepo := postgres.NewKnowledgeRetrievalRepository(db)
	httpClient := &http.Client{Timeout: 30 * time.Second}
	userSvc := service.NewUserService(userRepo, jwt, sessionStore)
	courseSvc := service.NewCourseService(courseRepo)
	chapterSvc := service.NewChapterService(chapterRepo, sectionRepo)
	sectionSvc := service.NewSectionService(sectionRepo)
	resourceSvc := service.NewResourceService(resourceRepo, sectionRepo, minioClient)
	summarySvc := service.NewResourceAISummaryService(resourceParseSvc, resourceRepo, teacherAIModelRepo, httpClient)
	enrollmentSvc := service.NewEnrollmentService(courseRepo, enrollmentRepo)
	catalogSvc := service.NewCatalogService(catalogRepo)
	progressSvc := service.NewProgressService(progressRepo)
	knowledgeSvc := service.NewKnowledgeService(embeddingChunkRepo, resourceParseSvc, teacherAIModelRepo, httpClient)
	aiModelSvc := service.NewAIModelService(teacherAIModelRepo, httpClient)
	knowledgeChatSvc := service.NewKnowledgeChatService(chatRepo, retrievalRepo, teacherAIModelRepo, cfg, httpClient)
	courseDesignSvc := service.NewCourseDesignService(cfg, courseRepo, teacherAIModelRepo, chapterSvc, sectionSvc, httpClient)
	questionBankSvc := service.NewQuestionBankService(cfg, courseRepo, teacherAIModelRepo, questionBankRepo, httpClient)
	lessonPlanSvc := service.NewSectionLessonPlanService(cfg, sectionRepo, resourceRepo, teacherAIModelRepo, httpClient)

	userH := handlers.NewUserHandler(userSvc)
	courseH := handlers.NewCourseHandler(courseSvc)
	chapterH := handlers.NewChapterHandler(chapterSvc)
	sectionH := handlers.NewSectionHandler(sectionSvc, lessonPlanSvc)
	resourceH := handlers.NewResourceHandler(resourceSvc, resourceParseSvc, summarySvc)
	enrollmentH := handlers.NewEnrollmentHandler(enrollmentSvc)
	catalogH := handlers.NewCatalogHandler(catalogSvc)
	progressH := handlers.NewProgressHandler(progressSvc)
	knowledgeH := handlers.NewKnowledgeHandler(knowledgeSvc)
	aiModelH := handlers.NewAIModelHandler(aiModelSvc)
	knowledgeChatH := handlers.NewKnowledgeChatHandler(knowledgeChatSvc)
	courseDesignH := handlers.NewCourseDesignHandler(courseDesignSvc)
	questionBankH := handlers.NewQuestionBankHandler(questionBankSvc)
	router := api.NewRouter(userH, courseH, chapterH, sectionH, resourceH, enrollmentH, catalogH, progressH, knowledgeH, aiModelH, knowledgeChatH, courseDesignH, questionBankH, jwt, sessionStore)

	ts := httptest.NewServer(router)
	defer ts.Close()

	suffix := time.Now().UnixNano()

	teacherAccess := registerAndLogin(t, ts.URL, "teacher", suffix)
	studentAccess := registerAndLogin(t, ts.URL, "student", suffix)

	createCourseResp := requestJSON(t, http.MethodPost, ts.URL+"/api/v1/teacher/courses", teacherAccess, map[string]any{
		"title":       "后端工程实战",
		"description": "课程简介",
	})
	if createCourseResp.StatusCode != http.StatusCreated {
		t.Fatalf("create course status=%d body=%s", createCourseResp.StatusCode, createCourseResp.BodyString)
	}
	courseID := asString(createCourseResp.JSON["id"])
	if courseID == "" {
		t.Fatalf("course id empty, body=%s", createCourseResp.BodyString)
	}

	updateCourseResp := requestJSON(t, http.MethodPut, ts.URL+"/api/v1/teacher/courses/"+courseID, teacherAccess, map[string]any{
		"title":       "后端工程实战",
		"description": "课程简介",
		"status":      "published",
	})
	if updateCourseResp.StatusCode != http.StatusNoContent {
		t.Fatalf("publish course status=%d body=%s", updateCourseResp.StatusCode, updateCourseResp.BodyString)
	}

	createChapterResp := requestJSON(t, http.MethodPost, ts.URL+"/api/v1/teacher/courses/"+courseID+"/chapters", teacherAccess, map[string]any{
		"title":      "第一章",
		"sort_order": 1,
	})
	if createChapterResp.StatusCode != http.StatusCreated {
		t.Fatalf("create chapter status=%d body=%s", createChapterResp.StatusCode, createChapterResp.BodyString)
	}
	chapterID := asString(createChapterResp.JSON["id"])
	if chapterID == "" {
		t.Fatalf("chapter id empty, body=%s", createChapterResp.BodyString)
	}

	sectionsResp := requestJSON(t, http.MethodGet, ts.URL+"/api/v1/teacher/courses/"+courseID+"/chapters/"+chapterID+"/sections", teacherAccess, nil)
	if sectionsResp.StatusCode != http.StatusOK {
		t.Fatalf("list sections status=%d body=%s", sectionsResp.StatusCode, sectionsResp.BodyString)
	}
	itemsAny, _ := sectionsResp.JSON["items"].([]any)
	if len(itemsAny) == 0 {
		t.Fatalf("expected default section, body=%s", sectionsResp.BodyString)
	}
	firstSec, _ := itemsAny[0].(map[string]any)
	sectionID := asString(firstSec["id"])
	if sectionID == "" {
		t.Fatalf("section id empty, body=%s", sectionsResp.BodyString)
	}

	initUploadResp := requestJSON(t, http.MethodPost, ts.URL+"/api/v1/teacher/sections/"+sectionID+"/resources/init-upload", teacherAccess, map[string]any{
		"course_id":     courseID,
		"file_name":     "lesson1.pdf",
		"resource_type": "pdf",
	})
	if initUploadResp.StatusCode != http.StatusOK {
		t.Fatalf("init upload status=%d body=%s", initUploadResp.StatusCode, initUploadResp.BodyString)
	}
	objectKey := asString(initUploadResp.JSON["object_key"])
	if objectKey == "" {
		t.Fatalf("object key empty, body=%s", initUploadResp.BodyString)
	}

	confirmResp := requestJSON(t, http.MethodPost, ts.URL+"/api/v1/teacher/sections/"+sectionID+"/resources/confirm", teacherAccess, map[string]any{
		"title":         "课程讲义",
		"resource_type": "pdf",
		"sort_order":    1,
		"object_key":    objectKey,
		"mime_type":     "application/pdf",
		"size_bytes":    1024,
	})
	if confirmResp.StatusCode != http.StatusCreated {
		t.Fatalf("confirm resource status=%d body=%s", confirmResp.StatusCode, confirmResp.BodyString)
	}
	resourceID := asString(confirmResp.JSON["id"])
	if resourceID == "" {
		t.Fatalf("resource id empty, body=%s", confirmResp.BodyString)
	}

	courseListResp := requestJSON(t, http.MethodGet, ts.URL+"/api/v1/courses", studentAccess, nil)
	if courseListResp.StatusCode != http.StatusOK {
		t.Fatalf("list courses status=%d body=%s", courseListResp.StatusCode, courseListResp.BodyString)
	}

	enrollResp := requestJSON(t, http.MethodPost, ts.URL+"/api/v1/courses/"+courseID+"/enroll", studentAccess, map[string]any{})
	if enrollResp.StatusCode != http.StatusNoContent {
		t.Fatalf("enroll status=%d body=%s", enrollResp.StatusCode, enrollResp.BodyString)
	}

	catalogResp := requestJSON(t, http.MethodGet, ts.URL+"/api/v1/my/courses/"+courseID+"/catalog", studentAccess, nil)
	if catalogResp.StatusCode != http.StatusOK {
		t.Fatalf("catalog status=%d body=%s", catalogResp.StatusCode, catalogResp.BodyString)
	}

	progressResp := requestJSON(t, http.MethodPost, ts.URL+"/api/v1/my/resources/"+resourceID+"/progress", studentAccess, map[string]any{
		"watched_seconds":  120,
		"progress_percent": 60,
	})
	if progressResp.StatusCode != http.StatusNoContent {
		t.Fatalf("progress status=%d body=%s", progressResp.StatusCode, progressResp.BodyString)
	}

	completeResp := requestJSON(t, http.MethodPost, ts.URL+"/api/v1/my/resources/"+resourceID+"/complete", studentAccess, map[string]any{})
	if completeResp.StatusCode != http.StatusNoContent {
		t.Fatalf("complete status=%d body=%s", completeResp.StatusCode, completeResp.BodyString)
	}
}

type jsonResp struct {
	StatusCode int
	BodyString string
	JSON       map[string]any
}

func requestJSON(t *testing.T, method, url, accessToken string, payload map[string]any) jsonResp {
	t.Helper()
	var bodyReader *bytes.Reader
	if payload == nil {
		bodyReader = bytes.NewReader(nil)
	} else {
		raw, _ := json.Marshal(payload)
		bodyReader = bytes.NewReader(raw)
	}
	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		t.Fatalf("new request failed: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if accessToken != "" {
		req.Header.Set("Authorization", "Bearer "+accessToken)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	buf := new(bytes.Buffer)
	_, _ = buf.ReadFrom(resp.Body)
	body := buf.String()

	out := map[string]any{}
	_ = json.Unmarshal(buf.Bytes(), &out)
	if data, ok := out["data"].(map[string]any); ok {
		for k, v := range data {
			if _, exists := out[k]; !exists {
				out[k] = v
			}
		}
	}
	return jsonResp{
		StatusCode: resp.StatusCode,
		BodyString: body,
		JSON:       out,
	}
}

func runMigration(ctx context.Context, db *gorm.DB) error {
	sqlPath := filepath.Join("..", "..", "migrations", "0001_init.sql")
	raw, err := os.ReadFile(sqlPath)
	if err != nil {
		return err
	}
	if err = db.WithContext(ctx).Exec(string(raw)).Error; err != nil {
		return err
	}
	// 集成测试基线使用 0001_init.sql；此处叠加与当前路由一致的结构变更（节层级与 AI 摘要列）。
	for _, name := range []string{
		"000009_course_sections.up.sql",
		"000010_resource_ai_summary.up.sql",
		"000012_course_question_bank.up.sql",
	} {
		p := filepath.Join("..", "..", "migrations", name)
		b, rerr := os.ReadFile(p)
		if rerr != nil {
			return rerr
		}
		if err = db.WithContext(ctx).Exec(string(b)).Error; err != nil {
			return err
		}
	}
	return nil
}

func flushRedis(ctx context.Context, cfg config.Config) error {
	sessionStore := authSvc.NewSessionStore(cfg)
	defer sessionStore.Close()
	return sessionStore.Flush(ctx)
}

func itoa(v int64) string {
	return strconv.FormatInt(v, 10)
}

func registerAndLogin(t *testing.T, baseURL, role string, suffix int64) string {
	t.Helper()
	account := role + itoa(suffix)
	registerBody := map[string]any{
		"email":    account + "@example.com",
		"username": account,
		"name":     role + "用户",
		"password": "Passw0rd!",
		"role":     role,
	}
	registerResp := requestJSON(t, http.MethodPost, baseURL+"/api/v1/auth/register", "", registerBody)
	if registerResp.StatusCode != http.StatusCreated {
		t.Fatalf("register(%s) status=%d body=%s", role, registerResp.StatusCode, registerResp.BodyString)
	}

	loginResp := requestJSON(t, http.MethodPost, baseURL+"/api/v1/auth/login", "", map[string]any{
		"account":  account,
		"password": "Passw0rd!",
	})
	if loginResp.StatusCode != http.StatusOK {
		t.Fatalf("login(%s) status=%d body=%s", role, loginResp.StatusCode, loginResp.BodyString)
	}
	return asString(loginResp.JSON["access_token"])
}

func asString(v any) string {
	if v == nil {
		return ""
	}
	switch x := v.(type) {
	case string:
		return x
	case float64:
		return strconv.FormatInt(int64(x), 10)
	case int64:
		return strconv.FormatInt(x, 10)
	default:
		return ""
	}
}
