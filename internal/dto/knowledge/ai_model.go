package knowledge

// AIModelListItem 列表展示用模型（不含 api_key）。
type AIModelListItem struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	ModelType  string `json:"model_type"`
	APIBaseURL string `json:"api_base_url"`
	ModelID    string `json:"model_id"`
	HasAPIKey  bool   `json:"has_api_key"`
}

// CreateAIModelReq 新建模型配置。
type CreateAIModelReq struct {
	Name       string `json:"name" binding:"required,max=128"`
	ModelType  string `json:"model_type" binding:"required,oneof=qa embedding rerank"`
	APIBaseURL string `json:"api_base_url" binding:"required"`
	ModelID    string `json:"model_id" binding:"required,max=256"`
	APIKey     string `json:"api_key"`
}

// CreateAIModelResp 创建成功返回 id。
type CreateAIModelResp struct {
	ID string `json:"id"`
}

// UpdateAIModelReq 更新模型；api_key 为空字符串表示不修改已存密钥。
type UpdateAIModelReq struct {
	Name       string `json:"name" binding:"required,max=128"`
	APIBaseURL string `json:"api_base_url" binding:"required"`
	ModelID    string `json:"model_id" binding:"required,max=256"`
	APIKey     string `json:"api_key"`
}

// TestAIModelConnectionReq 连通性探测入参（不落库）；api_key 为空且传 existing_model_id 时使用该条已存密钥。
type TestAIModelConnectionReq struct {
	ModelType         string `json:"model_type" binding:"required,oneof=qa embedding rerank"`
	APIBaseURL        string `json:"api_base_url" binding:"required"`
	ModelID           string `json:"model_id" binding:"required,max=256"`
	APIKey            string `json:"api_key"`
	ExistingModelID   *int64 `json:"existing_model_id,omitempty"`
}

// TestAIModelConnectionResp 探测结果；HTTP 2xx 且上游无 error 字段时 ok=true。
type TestAIModelConnectionResp struct {
	OK         bool   `json:"ok"`
	Message    string `json:"message"`
	HTTPStatus int    `json:"http_status,omitempty"`
}
