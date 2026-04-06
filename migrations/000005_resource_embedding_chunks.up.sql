-- resource_embedding_chunks：资源 RAG 分块 + 可选延迟嵌入（先确认分块，再写入向量）。
-- 不在库层声明外键。
-- embedding 使用 pgvector 的 vector 类型且不写维度修饰符：各行可不同维度；无法在列上建 ivfflat/hnsw（需固定维或按维度分表后再建索引）。

CREATE EXTENSION IF NOT EXISTS vector;

CREATE TABLE IF NOT EXISTS resource_embedding_chunks (
    id BIGSERIAL PRIMARY KEY,
    resource_id BIGINT NOT NULL,
    chunk_index INT NOT NULL CHECK (chunk_index >= 0),
    content TEXT NOT NULL,
    char_start INT CHECK (char_start IS NULL OR char_start >= 0),
    char_end INT CHECK (char_end IS NULL OR char_start IS NULL OR char_end >= char_start),
    token_count INT CHECK (token_count IS NULL OR token_count >= 0),
    metadata_json JSONB NOT NULL DEFAULT '{}'::jsonb,
    confirmed_at TIMESTAMPTZ,
    embedded_at TIMESTAMPTZ,
    -- 未嵌入前为 NULL；维度由向量自身表达，可用 vector_dims(embedding) 查询。
    embedding vector CHECK (
        embedding IS NULL
        OR (vector_dims(embedding) >= 1 AND vector_dims(embedding) <= 16000)
    ),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (resource_id, chunk_index)
);

CREATE INDEX IF NOT EXISTS idx_rec_chunks_resource_id ON resource_embedding_chunks (resource_id);
CREATE INDEX IF NOT EXISTS idx_rec_chunks_resource_confirmed ON resource_embedding_chunks (resource_id)
    WHERE confirmed_at IS NOT NULL AND embedded_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_rec_chunks_resource_embedded ON resource_embedding_chunks (resource_id)
    WHERE embedded_at IS NOT NULL;

COMMENT ON TABLE resource_embedding_chunks IS '课程资源分块与嵌入向量；无库级外键，resource_id 由应用保证指向 chapter_resources。';
COMMENT ON COLUMN resource_embedding_chunks.embedding IS 'pgvector vector（无 typmod）；维度以 vector_dims(embedding) 为准。该列不支持 ivfflat/hnsw。';
