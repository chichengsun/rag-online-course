-- 回滚：恢复 resource_parse_results（与 000004 一致；若曾写入 embedding_chunks 需自行处理数据）。
CREATE TABLE IF NOT EXISTS resource_parse_results (
    id BIGSERIAL PRIMARY KEY,
    resource_id BIGINT NOT NULL UNIQUE,
    status VARCHAR(32) NOT NULL,
    markdown TEXT,
    images_json JSONB NOT NULL DEFAULT '[]'::jsonb,
    metadata_json JSONB NOT NULL DEFAULT '{}'::jsonb,
    error_message TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_resource_parse_results_resource_id ON resource_parse_results (resource_id);
