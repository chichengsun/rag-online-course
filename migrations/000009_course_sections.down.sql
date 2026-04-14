-- 回滚「节」层级：删除 section 列与表。若曾存在多节合并，(chapter_id, sort_order) 可能无法恢复唯一性，仅适用于开发环境空库/可丢弃数据场景。

BEGIN;

DROP INDEX IF EXISTS idx_resources_section_id;
DROP INDEX IF EXISTS uk_chapter_resources_section_title;
DROP INDEX IF EXISTS uk_chapter_resources_section_sort;

ALTER TABLE chapter_resources DROP CONSTRAINT IF EXISTS chapter_resources_section_id_fkey;
ALTER TABLE chapter_resources DROP COLUMN IF EXISTS section_id;

CREATE UNIQUE INDEX uk_chapter_resources_course_chapter_title ON chapter_resources(course_id, chapter_id, title);
ALTER TABLE chapter_resources ADD CONSTRAINT chapter_resources_chapter_id_sort_order_key UNIQUE (chapter_id, sort_order);

DROP TABLE IF EXISTS course_sections;

COMMIT;
