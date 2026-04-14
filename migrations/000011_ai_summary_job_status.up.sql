-- AI 摘要异步任务状态：running / succeeded / failed / idle（未触发或历史数据默认）。

ALTER TABLE chapter_resources
  ADD COLUMN IF NOT EXISTS ai_summary_status VARCHAR(32) NOT NULL DEFAULT 'idle';

ALTER TABLE chapter_resources
  ADD COLUMN IF NOT EXISTS ai_summary_error TEXT;

-- 已有摘要的行标记为 succeeded，便于前端展示。
UPDATE chapter_resources
SET ai_summary_status = 'succeeded'
WHERE ai_summary IS NOT NULL AND TRIM(ai_summary) <> '' AND ai_summary_status = 'idle';
