-- 文档资源 AI 摘要（教师生成后可供列表/预览展示）。

ALTER TABLE chapter_resources ADD COLUMN IF NOT EXISTS ai_summary TEXT;
ALTER TABLE chapter_resources ADD COLUMN IF NOT EXISTS ai_summary_updated_at TIMESTAMPTZ;
