-- teacher_ai_models：教师配置的问答 / 嵌入 / 重排模型（OpenAI 兼容 HTTP）；库层不声明外键。
CREATE TABLE IF NOT EXISTS teacher_ai_models (
    id BIGSERIAL PRIMARY KEY,
    teacher_id BIGINT NOT NULL,
    name VARCHAR(128) NOT NULL,
    model_type VARCHAR(32) NOT NULL CHECK (model_type IN ('qa', 'embedding', 'rerank')),
    -- 嵌入/对话请求的完整 URL，例如 https://api.openai.com/v1/embeddings
    api_base_url TEXT NOT NULL,
    -- 请求体中的 model 字段值
    model_id VARCHAR(256) NOT NULL,
    api_key TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_teacher_ai_models_teacher ON teacher_ai_models (teacher_id);
CREATE INDEX IF NOT EXISTS idx_teacher_ai_models_teacher_type ON teacher_ai_models (teacher_id, model_type);

COMMENT ON TABLE teacher_ai_models IS '教师侧 AI 模型配置；api_key 仅存应用层，生产建议加密或改密钥托管。';
