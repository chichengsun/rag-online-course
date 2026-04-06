-- resource_parse_results：教师触发解析后的 Markdown / 图片元数据落库，供后续 RAG 切块索引。
CREATE TABLE IF NOT EXISTS resource_parse_results (
    id BIGSERIAL PRIMARY KEY,
    resource_id BIGINT NOT NULL UNIQUE REFERENCES chapter_resources (id) ON DELETE CASCADE,
    status VARCHAR(32) NOT NULL,
    markdown TEXT,
    images_json JSONB NOT NULL DEFAULT '[]'::jsonb,
    metadata_json JSONB NOT NULL DEFAULT '{}'::jsonb,
    error_message TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_resource_parse_results_resource_id ON resource_parse_results (resource_id);
