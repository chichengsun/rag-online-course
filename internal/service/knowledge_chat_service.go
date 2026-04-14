package service

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"rag-online-course/internal/config"
	"rag-online-course/internal/dto/knowledge"
	"rag-online-course/internal/logging"
	"rag-online-course/internal/repository/postgres"

	"github.com/sirupsen/logrus"
)

// KnowledgeChatService 编排知识库对话：检索、可选重排、问答生成与会话持久化。
type KnowledgeChatService struct {
	chatRepo   *postgres.KnowledgeChatRepository
	retrieval  *postgres.KnowledgeRetrievalRepository
	modelRepo  *postgres.TeacherAIModelRepository
	httpClient *http.Client
	rerankSvc  *RerankService
	ragCfg     config.RAGConfig
}

// ChatStreamCallbacks 定义流式问答过程中的事件回调。
type ChatStreamCallbacks struct {
	OnToken      func(token string)
	OnReferences func(refs []knowledge.ReferenceItem)
	OnDone       func(userMessageID, assistantMessageID string)
}

const (
	chatIntentSimple = "simple"
	chatIntentRAG    = "rag"
	chatHistoryLimit = 20
)

// ChatHistoryItem 表示会话内可供模型使用的一条历史消息。
type ChatHistoryItem struct {
	Role    string
	Content string
}

// NewKnowledgeChatService 创建知识库对话服务。
func NewKnowledgeChatService(
	chatRepo *postgres.KnowledgeChatRepository,
	retrieval *postgres.KnowledgeRetrievalRepository,
	modelRepo *postgres.TeacherAIModelRepository,
	cfg config.Config,
	httpClient *http.Client,
) *KnowledgeChatService {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	s := &KnowledgeChatService{
		chatRepo:   chatRepo,
		retrieval:  retrieval,
		modelRepo:  modelRepo,
		httpClient: httpClient,
		ragCfg:     cfg.RAG,
	}
	s.rerankSvc = NewRerankService(s)
	return s
}

func clampPage(page, pageSize int) (int, int) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}
	return page, pageSize
}

// CreateSession 创建会话。
func (s *KnowledgeChatService) CreateSession(ctx context.Context, teacherID, courseID int64, req knowledge.CreateChatSessionReq) (string, error) {
	title := strings.TrimSpace(req.Title)
	if title == "" {
		title = "新对话"
	}
	id, err := s.chatRepo.CreateSession(ctx, teacherID, courseID, truncateTitle(title))
	if err != nil {
		return "", err
	}
	return strconv.FormatInt(id, 10), nil
}

// ListSessions 会话分页列表。
func (s *KnowledgeChatService) ListSessions(ctx context.Context, teacherID int64, courseID *int64, page, pageSize int) (*knowledge.ListChatSessionsResp, error) {
	page, pageSize = clampPage(page, pageSize)
	offset := (page - 1) * pageSize
	items, total, err := s.chatRepo.ListSessions(ctx, teacherID, courseID, offset, pageSize)
	if err != nil {
		return nil, err
	}
	return &knowledge.ListChatSessionsResp{
		Page:     page,
		PageSize: pageSize,
		Total:    total,
		Items:    items,
	}, nil
}

// ListMessages 消息分页列表。
func (s *KnowledgeChatService) ListMessages(ctx context.Context, teacherID, sessionID int64, page, pageSize int) (*knowledge.ListChatMessagesResp, error) {
	page, pageSize = clampPage(page, pageSize)
	offset := (page - 1) * pageSize
	items, total, err := s.chatRepo.ListMessages(ctx, sessionID, teacherID, offset, pageSize)
	if err != nil {
		return nil, err
	}
	return &knowledge.ListChatMessagesResp{
		Page:     page,
		PageSize: pageSize,
		Total:    total,
		Items:    items,
	}, nil
}

// UpdateSessionTitle 更新会话标题。
func (s *KnowledgeChatService) UpdateSessionTitle(ctx context.Context, teacherID, sessionID int64, title string) error {
	t := truncateTitle(title)
	if strings.TrimSpace(t) == "" {
		return fmt.Errorf("标题不能为空")
	}
	return s.chatRepo.UpdateSessionTitle(ctx, sessionID, teacherID, t)
}

// DeleteSession 删除会话。
func (s *KnowledgeChatService) DeleteSession(ctx context.Context, teacherID, sessionID int64) error {
	return s.chatRepo.DeleteSession(ctx, sessionID, teacherID)
}

// AskInSession 在会话中继续问答并返回答案与引用。
func (s *KnowledgeChatService) AskInSession(ctx context.Context, teacherID, sessionID int64, req knowledge.AskInSessionReq) (*knowledge.AskInSessionResp, error) {
	start := time.Now()
	logger := logging.FromContext(ctx).WithField("session_id", sessionID)
	session, err := s.chatRepo.GetSession(ctx, sessionID, teacherID)
	if err != nil {
		return nil, err
	}
	courseID, _ := strconv.ParseInt(toString(session["course_id"]), 10, 64)
	question := strings.TrimSpace(req.Question)
	if question == "" {
		return nil, fmt.Errorf("问题不能为空")
	}
	history, err := s.loadRecentHistory(ctx, teacherID, sessionID, chatHistoryLimit)
	if err != nil {
		return nil, err
	}
	logger.WithFields(map[string]any{
		"course_id":     courseID,
		"question_len":  len([]rune(question)),
		"question":      truncateForLog(question, 120),
		"history_count": len(history),
		"top_k":         req.TopK,
		"semantic_min_score": func() float64 {
			if req.SemanticMinScore == nil {
				return 0
			}
			return *req.SemanticMinScore
		}(),
		"keyword_min_score": func() float64 {
			if req.KeywordMinScore == nil {
				return 0
			}
			return *req.KeywordMinScore
		}(),
		"use_rerank": req.UseRerank == nil || *req.UseRerank,
	}).Info("knowledge chat ask started")
	qaModel, embeddingModel, rerankModel, err := s.pickModels(ctx, teacherID, req.QAModelID)
	if err != nil {
		return nil, err
	}
	if qaModel == nil || strings.TrimSpace(qaModel.APIKey) == "" {
		return nil, fmt.Errorf("未配置可用问答模型（qa）")
	}
	intent, err := s.detectIntent(ctx, qaModel, question, history)
	if err != nil {
		logger.WithError(err).Warn("intent detect failed, fallback rag")
		intent = chatIntentRAG
	}
	logger.WithField("intent", intent).Info("knowledge chat intent decided")
	logger.WithFields(map[string]any{
		"qa_model_id":         qaModel.ID,
		"qa_model_name":       qaModel.Name,
		"qa_model_identifier": qaModel.ModelID,
	}).Info("knowledge chat qa model selected")
	var (
		answer     string
		references []knowledge.ReferenceItem
	)
	if intent == chatIntentSimple {
		answer, err = s.askQADirect(ctx, qaModel, question, history)
		if err != nil {
			return nil, err
		}
	} else {
		answer, references, err = s.runRAGPipeline(ctx, logger, teacherID, courseID, question, history, req, qaModel, embeddingModel, rerankModel)
		if err != nil {
			return nil, err
		}
	}
	modelSnapshot := map[string]any{
		"intent":    intent,
		"qa":        modelSnapshot(qaModel),
		"embedding": modelSnapshot(embeddingModel),
		"rerank":    modelSnapshot(rerankModel),
	}
	refMaps := make([]map[string]any, 0, len(references))
	for _, r := range references {
		refMaps = append(refMaps, map[string]any{
			"citation_no":    r.CitationNo,
			"chunk_id":       r.ChunkID,
			"resource_id":    r.ResourceID,
			"resource_title": r.ResourceTitle,
			"chunk_index":    r.ChunkIndex,
			"score":          r.Score,
			"snippet":        r.Snippet,
			"full_content":   r.FullContent,
		})
	}
	userMsgID, err := s.chatRepo.InsertMessage(ctx, sessionID, teacherID, "user", question, nil, modelSnapshot)
	if err != nil {
		return nil, err
	}
	assistantMsgID, err := s.chatRepo.InsertMessage(ctx, sessionID, teacherID, "assistant", answer, refMaps, modelSnapshot)
	if err != nil {
		return nil, err
	}
	if err := s.chatRepo.TouchSession(ctx, sessionID, teacherID); err != nil {
		logger.WithError(err).Warn("touch session failed")
	}
	logger.WithFields(map[string]any{
		"course_id":      courseID,
		"intent":         intent,
		"retrieval_size": len(references),
		"cost_ms":        time.Since(start).Milliseconds(),
	}).Info("knowledge chat ask success")
	return &knowledge.AskInSessionResp{
		SessionID:          strconv.FormatInt(sessionID, 10),
		UserMessageID:      strconv.FormatInt(userMsgID, 10),
		AssistantMessageID: strconv.FormatInt(assistantMsgID, 10),
		Answer:             answer,
		References:         references,
	}, nil
}

// AskInSessionStream 以流式方式返回答案 token，并在结束后持久化助手消息。
func (s *KnowledgeChatService) AskInSessionStream(ctx context.Context, teacherID, sessionID int64, req knowledge.AskInSessionReq, cb ChatStreamCallbacks) error {
	start := time.Now()
	logger := logging.FromContext(ctx).WithField("session_id", sessionID)
	session, err := s.chatRepo.GetSession(ctx, sessionID, teacherID)
	if err != nil {
		return err
	}
	courseID, _ := strconv.ParseInt(toString(session["course_id"]), 10, 64)
	question := strings.TrimSpace(req.Question)
	if question == "" {
		return fmt.Errorf("问题不能为空")
	}
	history, err := s.loadRecentHistory(ctx, teacherID, sessionID, chatHistoryLimit)
	if err != nil {
		return err
	}
	logger.WithFields(map[string]any{
		"course_id":     courseID,
		"question_len":  len([]rune(question)),
		"question":      truncateForLog(question, 120),
		"history_count": len(history),
		"top_k":         req.TopK,
		"semantic_min_score": func() float64 {
			if req.SemanticMinScore == nil {
				return 0
			}
			return *req.SemanticMinScore
		}(),
		"keyword_min_score": func() float64 {
			if req.KeywordMinScore == nil {
				return 0
			}
			return *req.KeywordMinScore
		}(),
		"use_rerank": req.UseRerank == nil || *req.UseRerank,
	}).Info("knowledge chat stream started")
	qaModel, embeddingModel, rerankModel, err := s.pickModels(ctx, teacherID, req.QAModelID)
	if err != nil {
		return err
	}
	if qaModel == nil || strings.TrimSpace(qaModel.APIKey) == "" {
		return fmt.Errorf("未配置可用问答模型（qa）")
	}
	intent, err := s.detectIntent(ctx, qaModel, question, history)
	if err != nil {
		logger.WithError(err).Warn("intent detect failed, fallback rag")
		intent = chatIntentRAG
	}
	logger.WithField("intent", intent).Info("knowledge chat stream intent decided")
	logger.WithFields(map[string]any{
		"qa_model_id":         qaModel.ID,
		"qa_model_name":       qaModel.Name,
		"qa_model_identifier": qaModel.ModelID,
	}).Info("knowledge chat stream qa model selected")
	modelSnapshot := map[string]any{
		"intent":    intent,
		"qa":        modelSnapshot(qaModel),
		"embedding": modelSnapshot(embeddingModel),
		"rerank":    modelSnapshot(rerankModel),
	}
	userMsgID, err := s.chatRepo.InsertMessage(ctx, sessionID, teacherID, "user", question, nil, modelSnapshot)
	if err != nil {
		return err
	}
	var (
		answer     string
		references []knowledge.ReferenceItem
	)
	if intent == chatIntentSimple {
		answer, err = s.askQADirectStream(ctx, qaModel, question, history, cb.OnToken)
		if err != nil {
			return err
		}
	} else {
		answer, references, err = s.runRAGPipelineStream(ctx, logger, teacherID, courseID, question, history, req, qaModel, embeddingModel, rerankModel, cb.OnToken)
		if err != nil {
			return err
		}
	}
	if cb.OnReferences != nil {
		cb.OnReferences(references)
	}
	refMaps := make([]map[string]any, 0, len(references))
	for _, r := range references {
		refMaps = append(refMaps, map[string]any{
			"citation_no":    r.CitationNo,
			"chunk_id":       r.ChunkID,
			"resource_id":    r.ResourceID,
			"resource_title": r.ResourceTitle,
			"chunk_index":    r.ChunkIndex,
			"score":          r.Score,
			"snippet":        r.Snippet,
			"full_content":   r.FullContent,
		})
	}
	assistantMsgID, err := s.chatRepo.InsertMessage(ctx, sessionID, teacherID, "assistant", answer, refMaps, modelSnapshot)
	if err != nil {
		return err
	}
	_ = s.chatRepo.TouchSession(ctx, sessionID, teacherID)
	if cb.OnDone != nil {
		cb.OnDone(strconv.FormatInt(userMsgID, 10), strconv.FormatInt(assistantMsgID, 10))
	}
	logger.WithFields(map[string]any{
		"course_id":      courseID,
		"intent":         intent,
		"retrieval_size": len(references),
		"stream":         true,
		"cost_ms":        time.Since(start).Milliseconds(),
	}).Info("knowledge chat ask stream success")
	return nil
}

// pickModels 按教师选择 qa/embedding/rerank 模型；qa 支持显式指定。
func (s *KnowledgeChatService) pickModels(ctx context.Context, teacherID int64, selectedQAModelID *int64) (*postgres.TeacherAIModelRow, *postgres.TeacherAIModelRow, *postgres.TeacherAIModelRow, error) {
	models, err := s.modelRepo.ListByTeacher(ctx, teacherID)
	if err != nil {
		return nil, nil, nil, err
	}
	var qaModel, embeddingModel, rerankModel *postgres.TeacherAIModelRow
	for i := range models {
		m := models[i]
		switch m.ModelType {
		case "qa":
			if qaModel == nil {
				qaModel = &m
			}
		case "embedding":
			if embeddingModel == nil {
				embeddingModel = &m
			}
		case "rerank":
			if rerankModel == nil {
				rerankModel = &m
			}
		}
	}
	if selectedQAModelID != nil {
		row, gErr := s.modelRepo.GetByID(ctx, *selectedQAModelID, teacherID)
		if gErr != nil {
			return nil, nil, nil, fmt.Errorf("指定问答模型不存在或无权访问")
		}
		if row.ModelType != "qa" {
			return nil, nil, nil, fmt.Errorf("指定模型不是问答模型（qa）")
		}
		qaModel = row
	}
	return qaModel, embeddingModel, rerankModel, nil
}

func (s *KnowledgeChatService) runRAGPipeline(
	ctx context.Context,
	logger *logrus.Entry,
	teacherID, courseID int64,
	question string,
	history []ChatHistoryItem,
	req knowledge.AskInSessionReq,
	qaModel, embeddingModel, rerankModel *postgres.TeacherAIModelRow,
) (string, []knowledge.ReferenceItem, error) {
	ragStart := time.Now()
	if embeddingModel == nil || strings.TrimSpace(embeddingModel.APIKey) == "" {
		return "", nil, fmt.Errorf("未配置可用嵌入模型（embedding）")
	}
	rewriteStart := time.Now()
	rewriteQuestion, err := s.rewriteQuestion(ctx, qaModel, question, history)
	if err != nil {
		return "", nil, err
	}
	logger.WithFields(map[string]any{
		"rewrite_cost_ms": time.Since(rewriteStart).Milliseconds(),
		"rewrite_query":   truncateForLog(rewriteQuestion, 120),
	}).Info("rag rewrite finished")
	keywords, err := s.extractKeywords(ctx, qaModel, rewriteQuestion)
	if err != nil {
		logger.WithError(err).Warn("keyword extraction failed, fallback by rewritten query")
		keywords = fallbackKeywords(rewriteQuestion)
	}
	logger.WithFields(map[string]any{
		"keyword_count": len(keywords),
		"keywords":      keywords,
	}).Info("rag keywords prepared")
	retrieveStart := time.Now()
	topK := req.TopK
	if topK <= 0 {
		topK = 8
	}
	candidates, semanticCount, keywordCount, err := s.hybridRetrieveCandidates(
		ctx, teacherID, courseID, embeddingModel, rewriteQuestion, keywords, topK, req.SemanticMinScore, req.KeywordMinScore,
	)
	if err != nil {
		return "", nil, err
	}
	logger.WithFields(map[string]any{
		"retrieve_cost_ms": time.Since(retrieveStart).Milliseconds(),
		"retrieve_count":   len(candidates),
		"semantic_count":   semanticCount,
		"keyword_count":    keywordCount,
		"top_k":            topK,
		"retrieve_preview": summarizeCandidates(candidates, 3),
	}).Info("rag retrieve finished")
	if len(candidates) == 0 {
		fallback := s.getFallbackAnswer()
		logger.WithFields(map[string]any{
			"reason": "no_retrieval_candidates",
		}).Warn("rag fallback answer used")
		return fallback, []knowledge.ReferenceItem{}, nil
	}
	useRerank := req.UseRerank == nil || *req.UseRerank
	rerankApplied := false
	if useRerank && rerankModel != nil && strings.TrimSpace(rerankModel.APIKey) != "" {
		rerankStart := time.Now()
		reRanked, rErr := s.rerankSvc.RerankCandidates(ctx, rerankModel, question, candidates)
		if rErr != nil {
			logger.WithError(rErr).Warn("rerank failed, fallback to recall order")
		} else if len(reRanked) > 0 {
			candidates = reRanked
			rerankApplied = true
		}
		logger.WithFields(map[string]any{
			"rerank_cost_ms": time.Since(rerankStart).Milliseconds(),
			"rerank_applied": rerankApplied,
			"rerank_count":   len(candidates),
			"rerank_preview": summarizeCandidates(candidates, 3),
		}).Info("rag rerank finished")
	}
	references, contextBlocks := buildReferencesAndContext(candidates)
	generateStart := time.Now()
	answer, err := s.askQA(ctx, qaModel, question, history, contextBlocks)
	if err != nil {
		return "", nil, err
	}
	answer, normalizedRefs, err := normalizeAnswerAndReferences(answer, references)
	if err != nil {
		return "", nil, err
	}
	logger.WithFields(map[string]any{
		"generate_cost_ms":     time.Since(generateStart).Milliseconds(),
		"answer_len":           len([]rune(answer)),
		"answer_preview":       truncateForLog(answer, 200),
		"context_ref_count":    len(references),
		"normalized_ref_count": len(normalizedRefs),
		"normalized_refs":      summarizeReferences(normalizedRefs, 5),
		"rag_total_cost_ms":    time.Since(ragStart).Milliseconds(),
	}).Info("rag pipeline finished")
	return answer, normalizedRefs, nil
}

func (s *KnowledgeChatService) runRAGPipelineStream(
	ctx context.Context,
	logger *logrus.Entry,
	teacherID, courseID int64,
	question string,
	history []ChatHistoryItem,
	req knowledge.AskInSessionReq,
	qaModel, embeddingModel, rerankModel *postgres.TeacherAIModelRow,
	onToken func(token string),
) (string, []knowledge.ReferenceItem, error) {
	ragStart := time.Now()
	if embeddingModel == nil || strings.TrimSpace(embeddingModel.APIKey) == "" {
		return "", nil, fmt.Errorf("未配置可用嵌入模型（embedding）")
	}
	rewriteStart := time.Now()
	rewriteQuestion, err := s.rewriteQuestion(ctx, qaModel, question, history)
	if err != nil {
		return "", nil, err
	}
	logger.WithFields(map[string]any{
		"rewrite_cost_ms": time.Since(rewriteStart).Milliseconds(),
		"rewrite_query":   truncateForLog(rewriteQuestion, 120),
		"stream":          true,
	}).Info("rag rewrite finished")
	keywords, err := s.extractKeywords(ctx, qaModel, rewriteQuestion)
	if err != nil {
		logger.WithError(err).Warn("keyword extraction failed, fallback by rewritten query")
		keywords = fallbackKeywords(rewriteQuestion)
	}
	logger.WithFields(map[string]any{
		"keyword_count": len(keywords),
		"keywords":      keywords,
		"stream":        true,
	}).Info("rag keywords prepared")
	retrieveStart := time.Now()
	topK := req.TopK
	if topK <= 0 {
		topK = 8
	}
	candidates, semanticCount, keywordCount, err := s.hybridRetrieveCandidates(
		ctx, teacherID, courseID, embeddingModel, rewriteQuestion, keywords, topK, req.SemanticMinScore, req.KeywordMinScore,
	)
	if err != nil {
		return "", nil, err
	}
	logger.WithFields(map[string]any{
		"retrieve_cost_ms": time.Since(retrieveStart).Milliseconds(),
		"retrieve_count":   len(candidates),
		"semantic_count":   semanticCount,
		"keyword_count":    keywordCount,
		"top_k":            topK,
		"retrieve_preview": summarizeCandidates(candidates, 3),
		"stream":           true,
	}).Info("rag retrieve finished")
	if len(candidates) == 0 {
		fallback := s.getFallbackAnswer()
		logger.WithFields(map[string]any{
			"reason": "no_retrieval_candidates",
			"stream": true,
		}).Warn("rag fallback answer used")
		if onToken != nil {
			onToken(fallback)
		}
		return fallback, []knowledge.ReferenceItem{}, nil
	}
	useRerank := req.UseRerank == nil || *req.UseRerank
	rerankApplied := false
	if useRerank && rerankModel != nil && strings.TrimSpace(rerankModel.APIKey) != "" {
		rerankStart := time.Now()
		reRanked, rErr := s.rerankSvc.RerankCandidates(ctx, rerankModel, question, candidates)
		if rErr != nil {
			logger.WithError(rErr).Warn("rerank failed, fallback to recall order")
		} else if len(reRanked) > 0 {
			candidates = reRanked
			rerankApplied = true
		}
		logger.WithFields(map[string]any{
			"rerank_cost_ms": time.Since(rerankStart).Milliseconds(),
			"rerank_applied": rerankApplied,
			"rerank_count":   len(candidates),
			"rerank_preview": summarizeCandidates(candidates, 3),
			"stream":         true,
		}).Info("rag rerank finished")
	}
	references, contextBlocks := buildReferencesAndContext(candidates)
	generateStart := time.Now()
	answer, err := s.askQAStream(ctx, qaModel, question, history, contextBlocks, onToken)
	if err != nil {
		return "", nil, err
	}
	answer, normalizedRefs, err := normalizeAnswerAndReferences(answer, references)
	if err != nil {
		return "", nil, err
	}
	logger.WithFields(map[string]any{
		"generate_cost_ms":     time.Since(generateStart).Milliseconds(),
		"answer_len":           len([]rune(answer)),
		"answer_preview":       truncateForLog(answer, 200),
		"context_ref_count":    len(references),
		"normalized_ref_count": len(normalizedRefs),
		"normalized_refs":      summarizeReferences(normalizedRefs, 5),
		"rag_total_cost_ms":    time.Since(ragStart).Milliseconds(),
		"stream":               true,
	}).Info("rag pipeline finished")
	return answer, normalizedRefs, nil
}

// detectIntent 识别用户问题应走 simple 还是 rag 链路。
func (s *KnowledgeChatService) detectIntent(ctx context.Context, model *postgres.TeacherAIModelRow, question string, history []ChatHistoryItem) (string, error) {
	start := time.Now()
	systemPrompt := strings.TrimSpace(s.ragCfg.IntentSystemPrompt)
	if systemPrompt == "" {
		systemPrompt = "你是意图分类器。只输出 simple 或 rag。simple=通用常识/闲聊/无需课程资料；rag=依赖课程知识库内容才能回答。"
	}
	userPrompt := question
	if len(history) > 0 {
		userPrompt = fmt.Sprintf("历史对话：\n%s\n\n当前问题：%s", formatHistoryForPrompt(history), question)
	}
	body := map[string]any{
		"model": strings.TrimSpace(model.ModelID),
		"messages": []map[string]string{
			{"role": "system", "content": systemPrompt},
			{"role": "user", "content": userPrompt},
		},
		"temperature": 0,
	}
	status, respBody, err := s.PostJSON(ctx, strings.TrimSpace(model.APIBaseURL), strings.TrimSpace(model.APIKey), body)
	if err != nil {
		return "", err
	}
	if status < 200 || status >= 300 {
		return "", fmt.Errorf("intent http %d: %s", status, truncateForLog(string(respBody), 300))
	}
	var parsed struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(respBody, &parsed); err != nil {
		return "", err
	}
	if len(parsed.Choices) == 0 {
		return "", fmt.Errorf("intent empty")
	}
	out := strings.ToLower(strings.TrimSpace(parsed.Choices[0].Message.Content))
	if strings.Contains(out, chatIntentSimple) {
		logging.FromContext(ctx).WithFields(map[string]any{
			"intent":         chatIntentSimple,
			"intent_raw":     truncateForLog(out, 80),
			"intent_cost_ms": time.Since(start).Milliseconds(),
			"history_count":  len(history),
		}).Info("intent detect success")
		return chatIntentSimple, nil
	}
	if strings.Contains(out, chatIntentRAG) {
		logging.FromContext(ctx).WithFields(map[string]any{
			"intent":         chatIntentRAG,
			"intent_raw":     truncateForLog(out, 80),
			"intent_cost_ms": time.Since(start).Milliseconds(),
			"history_count":  len(history),
		}).Info("intent detect success")
		return chatIntentRAG, nil
	}
	logging.FromContext(ctx).WithFields(map[string]any{
		"intent":         chatIntentRAG,
		"intent_raw":     truncateForLog(out, 80),
		"intent_cost_ms": time.Since(start).Milliseconds(),
	}).Info("intent detect fallback rag")
	return chatIntentRAG, nil
}

// rewriteQuestion 对 RAG 检索查询进行改写。
func (s *KnowledgeChatService) rewriteQuestion(ctx context.Context, model *postgres.TeacherAIModelRow, question string, history []ChatHistoryItem) (string, error) {
	start := time.Now()
	systemPrompt := strings.TrimSpace(s.ragCfg.RewriteSystemPrompt)
	if systemPrompt == "" {
		systemPrompt = "你是检索查询改写器。请把用户问题改写成便于语义检索的一句话，保留关键实体与约束。只输出改写结果。"
	}
	userPrompt := question
	if len(history) > 0 {
		userPrompt = fmt.Sprintf("历史对话：\n%s\n\n请改写当前问题：%s", formatHistoryForPrompt(history), question)
	}
	body := map[string]any{
		"model": strings.TrimSpace(model.ModelID),
		"messages": []map[string]string{
			{"role": "system", "content": systemPrompt},
			{"role": "user", "content": userPrompt},
		},
		"temperature": 0.1,
	}
	status, respBody, err := s.PostJSON(ctx, strings.TrimSpace(model.APIBaseURL), strings.TrimSpace(model.APIKey), body)
	if err != nil {
		return "", err
	}
	if status < 200 || status >= 300 {
		return "", fmt.Errorf("rewrite http %d: %s", status, truncateForLog(string(respBody), 300))
	}
	var parsed struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(respBody, &parsed); err != nil {
		return "", err
	}
	if len(parsed.Choices) == 0 {
		return question, nil
	}
	out := strings.TrimSpace(parsed.Choices[0].Message.Content)
	if out == "" {
		logging.FromContext(ctx).WithField("rewrite_cost_ms", time.Since(start).Milliseconds()).Info("rewrite empty, fallback original question")
		return question, nil
	}
	logging.FromContext(ctx).WithFields(map[string]any{
		"rewrite_cost_ms": time.Since(start).Milliseconds(),
		"rewrite_from":    truncateForLog(question, 120),
		"rewrite_query":   truncateForLog(out, 120),
	}).Info("rewrite success")
	return out, nil
}

func (s *KnowledgeChatService) askQADirect(ctx context.Context, model *postgres.TeacherAIModelRow, question string, history []ChatHistoryItem) (string, error) {
	systemPrompt := strings.TrimSpace(s.ragCfg.SimpleQASystemPrompt)
	if systemPrompt == "" {
		systemPrompt = "你是通用问答助手，请直接回答用户问题。"
	}
	msgs := buildChatMessages(systemPrompt, history, question)
	body := map[string]any{
		"model":       strings.TrimSpace(model.ModelID),
		"messages":    msgs,
		"temperature": 0.4,
	}
	status, respBody, err := s.PostJSON(ctx, strings.TrimSpace(model.APIBaseURL), strings.TrimSpace(model.APIKey), body)
	if err != nil {
		return "", err
	}
	if status < 200 || status >= 300 {
		return "", fmt.Errorf("qa direct http %d: %s", status, truncateForLog(string(respBody), 400))
	}
	var parsed struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(respBody, &parsed); err != nil {
		return "", fmt.Errorf("解析问答响应失败: %w", err)
	}
	if len(parsed.Choices) == 0 {
		return "", fmt.Errorf("问答模型未返回内容")
	}
	out := strings.TrimSpace(parsed.Choices[0].Message.Content)
	if out == "" {
		return "", fmt.Errorf("问答模型返回空内容")
	}
	return out, nil
}

func (s *KnowledgeChatService) askQADirectStream(ctx context.Context, model *postgres.TeacherAIModelRow, question string, history []ChatHistoryItem, onToken func(token string)) (string, error) {
	systemPrompt := strings.TrimSpace(s.ragCfg.SimpleQASystemPrompt)
	if systemPrompt == "" {
		systemPrompt = "你是通用问答助手，请直接回答用户问题。"
	}
	msgs := buildChatMessages(systemPrompt, history, question)
	body := map[string]any{
		"model":       strings.TrimSpace(model.ModelID),
		"messages":    msgs,
		"temperature": 0.4,
		"stream":      true,
	}
	raw, err := json.Marshal(body)
	if err != nil {
		return "", err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, strings.TrimSpace(model.APIBaseURL), bytes.NewReader(raw))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+strings.TrimSpace(model.APIKey))
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		b, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("qa stream http %d: %s", resp.StatusCode, truncateForLog(string(b), 400))
	}
	contentType := strings.ToLower(resp.Header.Get("Content-Type"))
	if !strings.Contains(contentType, "text/event-stream") {
		b, _ := io.ReadAll(resp.Body)
		var parsed struct {
			Choices []struct {
				Message struct {
					Content string `json:"content"`
				} `json:"message"`
			} `json:"choices"`
		}
		if err := json.Unmarshal(b, &parsed); err != nil {
			return "", fmt.Errorf("解析问答响应失败: %w", err)
		}
		if len(parsed.Choices) == 0 {
			return "", fmt.Errorf("问答模型未返回内容")
		}
		out := strings.TrimSpace(parsed.Choices[0].Message.Content)
		if onToken != nil && out != "" {
			onToken(out)
		}
		return out, nil
	}
	reader := bufio.NewReader(resp.Body)
	var answer strings.Builder
	for {
		line, rErr := reader.ReadString('\n')
		if rErr != nil && len(line) == 0 {
			break
		}
		line = strings.TrimSpace(line)
		if line == "" || !strings.HasPrefix(line, "data:") {
			if rErr != nil {
				break
			}
			continue
		}
		payload := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
		if payload == "[DONE]" {
			break
		}
		var chunk struct {
			Choices []struct {
				Delta struct {
					Content string `json:"content"`
				} `json:"delta"`
				Message struct {
					Content string `json:"content"`
				} `json:"message"`
			} `json:"choices"`
		}
		if err := json.Unmarshal([]byte(payload), &chunk); err != nil {
			if rErr != nil {
				break
			}
			continue
		}
		if len(chunk.Choices) == 0 {
			if rErr != nil {
				break
			}
			continue
		}
		token := chunk.Choices[0].Delta.Content
		if token == "" {
			token = chunk.Choices[0].Message.Content
		}
		if token != "" {
			answer.WriteString(token)
			if onToken != nil {
				onToken(token)
			}
		}
		if rErr != nil {
			break
		}
	}
	out := strings.TrimSpace(answer.String())
	if out == "" {
		return "", fmt.Errorf("问答模型返回空内容")
	}
	return out, nil
}

// PostJSON 实现 RerankService 的 HTTPJSONClient。
func (s *KnowledgeChatService) PostJSON(ctx context.Context, apiURL, apiKey string, body map[string]any) (int, []byte, error) {
	raw, err := json.Marshal(body)
	if err != nil {
		return 0, nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, apiURL, bytes.NewReader(raw))
	if err != nil {
		return 0, nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return 0, nil, err
	}
	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return resp.StatusCode, nil, err
	}
	return resp.StatusCode, b, nil
}

func (s *KnowledgeChatService) embedQuery(ctx context.Context, model *postgres.TeacherAIModelRow, query string) (string, error) {
	body := map[string]any{
		"model": strings.TrimSpace(model.ModelID),
		"input": query,
	}
	status, respBody, err := s.PostJSON(ctx, strings.TrimSpace(model.APIBaseURL), strings.TrimSpace(model.APIKey), body)
	if err != nil {
		return "", err
	}
	if status < 200 || status >= 300 {
		return "", fmt.Errorf("embedding http %d: %s", status, truncateForLog(string(respBody), 400))
	}
	var parsed struct {
		Data []struct {
			Embedding []float64 `json:"embedding"`
		} `json:"data"`
	}
	if err := json.Unmarshal(respBody, &parsed); err != nil {
		return "", fmt.Errorf("解析 embedding 响应失败: %w", err)
	}
	if len(parsed.Data) == 0 || len(parsed.Data[0].Embedding) == 0 {
		return "", fmt.Errorf("embedding 返回为空")
	}
	return floatsToVectorLiteral(parsed.Data[0].Embedding), nil
}

func (s *KnowledgeChatService) askQA(ctx context.Context, model *postgres.TeacherAIModelRow, question string, history []ChatHistoryItem, contextBlocks string) (string, error) {
	systemPrompt := strings.TrimSpace(s.ragCfg.QASystemPrompt)
	if systemPrompt == "" {
		systemPrompt = "你是课程知识库问答助手。你必须且只能依据给定的检索上下文回答。若上下文不足，请原样回复：{{fallback_answer}}。"
	}
	fallbackAnswer := s.getFallbackAnswer()
	systemPrompt = strings.ReplaceAll(systemPrompt, "{{fallback_answer}}", fallbackAnswer)
	userPrompt := fmt.Sprintf("问题：%s\n\n上下文：\n%s", question, contextBlocks)
	msgs := buildChatMessages(systemPrompt, history, userPrompt)
	body := map[string]any{
		"model":       strings.TrimSpace(model.ModelID),
		"messages":    msgs,
		"temperature": 0.2,
	}
	status, respBody, err := s.PostJSON(ctx, strings.TrimSpace(model.APIBaseURL), strings.TrimSpace(model.APIKey), body)
	if err != nil {
		return "", err
	}
	if status < 200 || status >= 300 {
		return "", fmt.Errorf("qa http %d: %s", status, truncateForLog(string(respBody), 400))
	}
	var parsed struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(respBody, &parsed); err != nil {
		return "", fmt.Errorf("解析问答响应失败: %w", err)
	}
	if len(parsed.Choices) == 0 {
		return "", fmt.Errorf("问答模型未返回内容")
	}
	out := strings.TrimSpace(parsed.Choices[0].Message.Content)
	if out == "" {
		return "", fmt.Errorf("问答模型返回空内容")
	}
	return out, nil
}

func (s *KnowledgeChatService) askQAStream(
	ctx context.Context,
	model *postgres.TeacherAIModelRow,
	question string,
	history []ChatHistoryItem,
	contextBlocks string,
	onToken func(token string),
) (string, error) {
	systemPrompt := strings.TrimSpace(s.ragCfg.QASystemPrompt)
	if systemPrompt == "" {
		systemPrompt = "你是课程知识库问答助手。你必须且只能依据给定的检索上下文回答。若上下文不足，请原样回复：{{fallback_answer}}。"
	}
	fallbackAnswer := s.getFallbackAnswer()
	systemPrompt = strings.ReplaceAll(systemPrompt, "{{fallback_answer}}", fallbackAnswer)
	userPrompt := fmt.Sprintf("问题：%s\n\n上下文：\n%s", question, contextBlocks)
	msgs := buildChatMessages(systemPrompt, history, userPrompt)
	body := map[string]any{
		"model":       strings.TrimSpace(model.ModelID),
		"messages":    msgs,
		"temperature": 0.2,
		"stream":      true,
	}
	raw, err := json.Marshal(body)
	if err != nil {
		return "", err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, strings.TrimSpace(model.APIBaseURL), bytes.NewReader(raw))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+strings.TrimSpace(model.APIKey))
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		b, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("qa stream http %d: %s", resp.StatusCode, truncateForLog(string(b), 400))
	}
	contentType := strings.ToLower(resp.Header.Get("Content-Type"))
	// 兼容不支持 stream 的上游：直接按普通 completion 解析。
	if !strings.Contains(contentType, "text/event-stream") {
		b, _ := io.ReadAll(resp.Body)
		var parsed struct {
			Choices []struct {
				Message struct {
					Content string `json:"content"`
				} `json:"message"`
			} `json:"choices"`
		}
		if err := json.Unmarshal(b, &parsed); err != nil {
			return "", fmt.Errorf("解析问答响应失败: %w", err)
		}
		if len(parsed.Choices) == 0 {
			return "", fmt.Errorf("问答模型未返回内容")
		}
		out := strings.TrimSpace(parsed.Choices[0].Message.Content)
		if onToken != nil && out != "" {
			onToken(out)
		}
		return out, nil
	}

	reader := bufio.NewReader(resp.Body)
	var answer strings.Builder
	for {
		line, rErr := reader.ReadString('\n')
		if rErr != nil && len(line) == 0 {
			break
		}
		line = strings.TrimSpace(line)
		if line == "" || !strings.HasPrefix(line, "data:") {
			if rErr != nil {
				break
			}
			continue
		}
		payload := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
		if payload == "[DONE]" {
			break
		}
		var chunk struct {
			Choices []struct {
				Delta struct {
					Content string `json:"content"`
				} `json:"delta"`
				Message struct {
					Content string `json:"content"`
				} `json:"message"`
			} `json:"choices"`
		}
		if err := json.Unmarshal([]byte(payload), &chunk); err != nil {
			if rErr != nil {
				break
			}
			continue
		}
		if len(chunk.Choices) == 0 {
			if rErr != nil {
				break
			}
			continue
		}
		token := chunk.Choices[0].Delta.Content
		if token == "" {
			token = chunk.Choices[0].Message.Content
		}
		if token != "" {
			answer.WriteString(token)
			if onToken != nil {
				onToken(token)
			}
		}
		if rErr != nil {
			break
		}
	}
	out := strings.TrimSpace(answer.String())
	if out == "" {
		return "", fmt.Errorf("问答模型返回空内容")
	}
	return out, nil
}

func (s *KnowledgeChatService) getFallbackAnswer() string {
	fallbackAnswer := strings.TrimSpace(s.ragCfg.FallbackAnswer)
	if fallbackAnswer == "" {
		return "我在当前课程知识库中找不到足够依据，请补充资料或换个问法。"
	}
	return fallbackAnswer
}

func truncateTitle(s string) string {
	rs := []rune(strings.TrimSpace(s))
	if len(rs) <= 60 {
		return string(rs)
	}
	return string(rs[:60])
}

func buildReferencesAndContext(candidates []postgres.RetrievalChunk) ([]knowledge.ReferenceItem, string) {
	refs := make([]knowledge.ReferenceItem, 0, len(candidates))
	var blocks []string
	for i, c := range candidates {
		snippet := c.Content
		if len([]rune(snippet)) > 200 {
			snippet = string([]rune(snippet)[:200]) + "..."
		}
		score := c.Score
		if score == 0 {
			score = 1 - c.Distance
		}
		ref := knowledge.ReferenceItem{
			CitationNo:    i + 1,
			ChunkID:       strconv.FormatInt(c.ChunkID, 10),
			ResourceID:    strconv.FormatInt(c.ResourceID, 10),
			ResourceTitle: c.ResourceTitle,
			ChunkIndex:    c.ChunkIndex,
			Score:         score,
			Snippet:       snippet,
			FullContent:   c.Content,
		}
		refs = append(refs, ref)
		blocks = append(blocks, fmt.Sprintf(
			"[%d] 资源:%s(#%d)\n%s",
			i+1, c.ResourceTitle, c.ChunkIndex+1, c.Content,
		))
	}
	return refs, strings.Join(blocks, "\n\n")
}

// normalizeAnswerAndReferences 统一回答中的引用编号，并仅保留实际引用项。
func normalizeAnswerAndReferences(answer string, refs []knowledge.ReferenceItem) (string, []knowledge.ReferenceItem, error) {
	if len(refs) == 0 {
		return stripCitationMarkers(answer), nil, nil
	}
	used := extractCitationNumbers(answer)
	// 若模型未显式输出编号，兜底追加前若干引用编号，保证“正文可点 + 底部可查阅”。
	if len(used) == 0 {
		fallbackCount := 3
		if len(refs) < fallbackCount {
			fallbackCount = len(refs)
		}
		selected := make([]knowledge.ReferenceItem, 0, fallbackCount)
		for i := 0; i < fallbackCount; i++ {
			selected = append(selected, refs[i])
		}
		return stripCitationMarkers(answer), selected, nil
	}

	refByOldNo := make(map[int]knowledge.ReferenceItem, len(refs))
	for _, r := range refs {
		refByOldNo[r.CitationNo] = r
	}
	normalized := make([]knowledge.ReferenceItem, 0, len(used))
	remap := make(map[int]int, len(used))
	newNo := 1
	for _, oldNo := range used {
		ref, ok := refByOldNo[oldNo]
		if !ok {
			continue
		}
		ref.CitationNo = newNo
		normalized = append(normalized, ref)
		remap[oldNo] = newNo
		newNo++
	}
	if len(normalized) == 0 {
		return answer, nil, nil
	}
	sort.Slice(normalized, func(i, j int) bool { return normalized[i].CitationNo < normalized[j].CitationNo })
	answer = rewriteCitationNumbers(answer, remap)
	return stripCitationMarkers(answer), normalized, nil
}

func extractCitationNumbers(answer string) []int {
	re := regexp.MustCompile(`\[(\d+)\]`)
	matches := re.FindAllStringSubmatch(answer, -1)
	if len(matches) == 0 {
		return nil
	}
	seen := map[int]struct{}{}
	out := make([]int, 0, len(matches))
	for _, m := range matches {
		n, err := strconv.Atoi(m[1])
		if err != nil || n <= 0 {
			continue
		}
		if _, ok := seen[n]; ok {
			continue
		}
		seen[n] = struct{}{}
		out = append(out, n)
	}
	return out
}

func rewriteCitationNumbers(answer string, remap map[int]int) string {
	re := regexp.MustCompile(`\[(\d+)\]`)
	return re.ReplaceAllStringFunc(answer, func(m string) string {
		sub := re.FindStringSubmatch(m)
		if len(sub) != 2 {
			return m
		}
		n, err := strconv.Atoi(sub[1])
		if err != nil {
			return m
		}
		if to, ok := remap[n]; ok {
			return fmt.Sprintf("[%d]", to)
		}
		return m
	})
}

func stripCitationMarkers(answer string) string {
	re := regexp.MustCompile(`\[(\d+)\]`)
	out := re.ReplaceAllString(answer, "")
	// 清理因为删除引用序号产生的多余空白。
	out = strings.ReplaceAll(out, "  ", " ")
	out = strings.ReplaceAll(out, "\n \n", "\n\n")
	return strings.TrimSpace(out)
}

// loadRecentHistory 拉取会话最近若干条历史消息，供模型保持上下文连续性。
func (s *KnowledgeChatService) loadRecentHistory(ctx context.Context, teacherID, sessionID int64, limit int) ([]ChatHistoryItem, error) {
	if limit <= 0 {
		return nil, nil
	}
	items, _, err := s.chatRepo.ListMessages(ctx, sessionID, teacherID, 0, limit)
	if err != nil {
		return nil, err
	}
	out := make([]ChatHistoryItem, 0, len(items))
	for _, it := range items {
		role := strings.TrimSpace(toString(it["role"]))
		content := strings.TrimSpace(toString(it["content"]))
		if content == "" {
			continue
		}
		if role != "user" && role != "assistant" {
			continue
		}
		out = append(out, ChatHistoryItem{
			Role:    role,
			Content: content,
		})
	}
	return out, nil
}

func buildChatMessages(systemPrompt string, history []ChatHistoryItem, userPrompt string) []map[string]string {
	msgs := make([]map[string]string, 0, 2+len(history))
	msgs = append(msgs, map[string]string{"role": "system", "content": systemPrompt})
	for _, h := range history {
		msgs = append(msgs, map[string]string{"role": h.Role, "content": h.Content})
	}
	msgs = append(msgs, map[string]string{"role": "user", "content": userPrompt})
	return msgs
}

func formatHistoryForPrompt(history []ChatHistoryItem) string {
	if len(history) == 0 {
		return ""
	}
	lines := make([]string, 0, len(history))
	for _, h := range history {
		role := "用户"
		if h.Role == "assistant" {
			role = "助手"
		}
		lines = append(lines, fmt.Sprintf("%s：%s", role, h.Content))
	}
	return strings.Join(lines, "\n")
}

func summarizeCandidates(candidates []postgres.RetrievalChunk, limit int) []map[string]any {
	if len(candidates) == 0 {
		return nil
	}
	if limit <= 0 || limit > len(candidates) {
		limit = len(candidates)
	}
	out := make([]map[string]any, 0, limit)
	for i := 0; i < limit; i++ {
		c := candidates[i]
		out = append(out, map[string]any{
			"rank":           i + 1,
			"chunk_id":       c.ChunkID,
			"resource_id":    c.ResourceID,
			"resource_title": truncateForLog(c.ResourceTitle, 40),
			"chunk_index":    c.ChunkIndex,
			"distance":       c.Distance,
			"score":          c.Score,
			"content":        truncateForLog(c.Content, 80),
		})
	}
	return out
}

func summarizeReferences(refs []knowledge.ReferenceItem, limit int) []map[string]any {
	if len(refs) == 0 {
		return nil
	}
	if limit <= 0 || limit > len(refs) {
		limit = len(refs)
	}
	out := make([]map[string]any, 0, limit)
	for i := 0; i < limit; i++ {
		r := refs[i]
		out = append(out, map[string]any{
			"citation_no":    r.CitationNo,
			"chunk_id":       r.ChunkID,
			"resource_id":    r.ResourceID,
			"resource_title": truncateForLog(r.ResourceTitle, 40),
			"chunk_index":    r.ChunkIndex,
			"score":          r.Score,
			"snippet":        truncateForLog(r.Snippet, 80),
		})
	}
	return out
}

// hybridRetrieveCandidates 执行语义检索与关键词检索，并用 RRF 做融合排序。
func (s *KnowledgeChatService) hybridRetrieveCandidates(
	ctx context.Context,
	teacherID, courseID int64,
	embeddingModel *postgres.TeacherAIModelRow,
	query string,
	keywords []string,
	topK int,
	semanticMinScore *float64,
	keywordMinScore *float64,
) ([]postgres.RetrievalChunk, int, int, error) {
	queryVec, err := s.embedQuery(ctx, embeddingModel, query)
	if err != nil {
		return nil, 0, 0, err
	}
	semanticCandidates, err := s.retrieval.SearchByCourse(ctx, teacherID, courseID, queryVec, topK)
	if err != nil {
		return nil, 0, 0, err
	}
	semanticCandidates = filterCandidatesByScore(semanticCandidates, clampScoreThreshold(semanticMinScore))
	keywordCandidates, err := s.retrieval.SearchByCourseKeywordTokens(ctx, teacherID, courseID, keywords, topK)
	if err != nil {
		return nil, 0, 0, err
	}
	keywordCandidates = filterCandidatesByScore(keywordCandidates, clampScoreThreshold(keywordMinScore))
	fused := fuseCandidatesByRRF(semanticCandidates, keywordCandidates, topK)
	return fused, len(semanticCandidates), len(keywordCandidates), nil
}

func clampScoreThreshold(v *float64) float64 {
	if v == nil {
		return 0
	}
	if *v < 0 {
		return 0
	}
	if *v > 1 {
		return 1
	}
	return *v
}

func filterCandidatesByScore(items []postgres.RetrievalChunk, minScore float64) []postgres.RetrievalChunk {
	if minScore <= 0 {
		return items
	}
	out := make([]postgres.RetrievalChunk, 0, len(items))
	for _, it := range items {
		if it.Score >= minScore {
			out = append(out, it)
		}
	}
	return out
}

// extractKeywords 对改写后的问题做分词关键词抽取。
func (s *KnowledgeChatService) extractKeywords(ctx context.Context, model *postgres.TeacherAIModelRow, rewritten string) ([]string, error) {
	systemPrompt := strings.TrimSpace(s.ragCfg.KeywordSystemPrompt)
	if systemPrompt == "" {
		systemPrompt = "你是检索关键词分解器。仅输出 JSON 数组关键词，例如：[\"goroutine\",\"调度\",\"GPM\"]"
	}
	body := map[string]any{
		"model": strings.TrimSpace(model.ModelID),
		"messages": []map[string]string{
			{"role": "system", "content": systemPrompt},
			{"role": "user", "content": rewritten},
		},
		"temperature": 0,
	}
	status, respBody, err := s.PostJSON(ctx, strings.TrimSpace(model.APIBaseURL), strings.TrimSpace(model.APIKey), body)
	if err != nil {
		return nil, err
	}
	if status < 200 || status >= 300 {
		return nil, fmt.Errorf("keyword extract http %d: %s", status, truncateForLog(string(respBody), 300))
	}
	var parsed struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(respBody, &parsed); err != nil {
		return nil, err
	}
	if len(parsed.Choices) == 0 {
		return nil, fmt.Errorf("keyword extract empty")
	}
	raw := strings.TrimSpace(parsed.Choices[0].Message.Content)
	keywords := parseKeywords(raw)
	if len(keywords) == 0 {
		return nil, fmt.Errorf("keyword extract parse empty")
	}
	return keywords, nil
}

func parseKeywords(raw string) []string {
	out := make([]string, 0)
	// 优先解析 JSON 数组
	var arr []string
	if err := json.Unmarshal([]byte(raw), &arr); err == nil {
		for _, it := range arr {
			v := strings.TrimSpace(it)
			if v != "" {
				out = append(out, v)
			}
		}
		return dedupeKeywords(out, 8)
	}
	// 兜底：按常见分隔符切分
	items := strings.FieldsFunc(raw, func(r rune) bool {
		return r == ',' || r == '，' || r == ';' || r == '；' || r == '\n' || r == '\t' || r == ' ' || r == '、'
	})
	for _, it := range items {
		v := strings.Trim(strings.TrimSpace(it), "\"'[]")
		if v != "" {
			out = append(out, v)
		}
	}
	return dedupeKeywords(out, 8)
}

func fallbackKeywords(q string) []string {
	items := strings.FieldsFunc(q, func(r rune) bool {
		return r == ',' || r == '，' || r == ';' || r == '；' || r == '\n' || r == '\t' || r == ' ' || r == '、'
	})
	return dedupeKeywords(items, 8)
}

func dedupeKeywords(items []string, limit int) []string {
	if limit <= 0 {
		limit = 8
	}
	seen := map[string]struct{}{}
	out := make([]string, 0, len(items))
	for _, it := range items {
		v := strings.TrimSpace(it)
		if v == "" {
			continue
		}
		k := strings.ToLower(v)
		if _, ok := seen[k]; ok {
			continue
		}
		seen[k] = struct{}{}
		out = append(out, v)
		if len(out) >= limit {
			break
		}
	}
	return out
}

// fuseCandidatesByRRF 使用倒数排名融合语义与关键词两路结果。
func fuseCandidatesByRRF(
	semantic []postgres.RetrievalChunk,
	keyword []postgres.RetrievalChunk,
	limit int,
) []postgres.RetrievalChunk {
	if limit <= 0 {
		limit = 10
	}
	type agg struct {
		chunk postgres.RetrievalChunk
		score float64
	}
	const k = 60.0
	m := make(map[int64]*agg, len(semantic)+len(keyword))
	apply := func(items []postgres.RetrievalChunk) {
		for i, c := range items {
			rank := float64(i + 1)
			rrf := 1.0 / (k + rank)
			if ex, ok := m[c.ChunkID]; ok {
				ex.score += rrf
				continue
			}
			cp := c
			m[c.ChunkID] = &agg{chunk: cp, score: rrf}
		}
	}
	apply(semantic)
	apply(keyword)
	merged := make([]agg, 0, len(m))
	for _, v := range m {
		// Distance 继续保留“越小越靠前”语义，这里用 1-score 近似映射。
		v.chunk.Distance = 1 - v.score
		v.chunk.Score = v.score
		merged = append(merged, *v)
	}
	sort.Slice(merged, func(i, j int) bool {
		if merged[i].score == merged[j].score {
			return merged[i].chunk.ChunkID > merged[j].chunk.ChunkID
		}
		return merged[i].score > merged[j].score
	})
	if len(merged) > limit {
		merged = merged[:limit]
	}
	out := make([]postgres.RetrievalChunk, 0, len(merged))
	for _, it := range merged {
		out = append(out, it.chunk)
	}
	return out
}

func modelSnapshot(m *postgres.TeacherAIModelRow) map[string]any {
	if m == nil {
		return nil
	}
	return map[string]any{
		"id":           m.ID,
		"name":         m.Name,
		"model_type":   m.ModelType,
		"api_base_url": m.APIBaseURL,
		"model_id":     m.ModelID,
	}
}
