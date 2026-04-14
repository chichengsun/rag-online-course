-- 知识库对话会话与消息：支撑分页会话列表、历史消息查询与继续问答；库层不声明外键。
CREATE TABLE IF NOT EXISTS knowledge_chat_sessions (
    id BIGSERIAL PRIMARY KEY,
    teacher_id BIGINT NOT NULL,
    course_id BIGINT NOT NULL,
    title VARCHAR(256) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_kchat_sessions_teacher ON knowledge_chat_sessions (teacher_id);
CREATE INDEX IF NOT EXISTS idx_kchat_sessions_teacher_course ON knowledge_chat_sessions (teacher_id, course_id);
CREATE INDEX IF NOT EXISTS idx_kchat_sessions_teacher_updated ON knowledge_chat_sessions (teacher_id, updated_at DESC);

CREATE TABLE IF NOT EXISTS knowledge_chat_messages (
    id BIGSERIAL PRIMARY KEY,
    session_id BIGINT NOT NULL,
    role VARCHAR(16) NOT NULL CHECK (role IN ('user', 'assistant', 'system')),
    content TEXT NOT NULL,
    -- 回答引用：资源/分块/分数/片段等，便于前端可追溯展示。
    references_json JSONB NOT NULL DEFAULT '[]'::jsonb,
    -- 记录当次调用的模型快照（qa/embedding/rerank）。
    model_snapshot_json JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_kchat_messages_session ON knowledge_chat_messages (session_id, created_at ASC);

COMMENT ON TABLE knowledge_chat_sessions IS '知识库对话会话（教师侧）。';
COMMENT ON TABLE knowledge_chat_messages IS '知识库对话消息（用户提问与助手回答）。';
