-- 在「章节」与「资源」之间增加 course_sections（节）；资源归属节，排序在节内唯一。
-- 历史数据：为每个章节插入「默认节」并将原资源挂到该节下。

BEGIN;

CREATE TABLE course_sections (
    id BIGSERIAL PRIMARY KEY,
    chapter_id BIGINT NOT NULL REFERENCES course_chapters(id) ON DELETE CASCADE,
    course_id BIGINT NOT NULL REFERENCES courses(id) ON DELETE CASCADE,
    title VARCHAR(255) NOT NULL,
    sort_order INT NOT NULL CHECK (sort_order > 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (chapter_id, sort_order)
);
CREATE INDEX idx_course_sections_chapter_id ON course_sections(chapter_id);
CREATE INDEX idx_course_sections_course_id ON course_sections(course_id);
CREATE UNIQUE INDEX uk_course_sections_chapter_title ON course_sections(chapter_id, title);

ALTER TABLE chapter_resources ADD COLUMN IF NOT EXISTS section_id BIGINT REFERENCES course_sections(id) ON DELETE CASCADE;

INSERT INTO course_sections (chapter_id, course_id, title, sort_order)
SELECT c.id, c.course_id, '默认节', 1
FROM course_chapters c
WHERE NOT EXISTS (SELECT 1 FROM course_sections s WHERE s.chapter_id = c.id);

UPDATE chapter_resources r
SET section_id = s.id
FROM course_sections s
WHERE r.section_id IS NULL
  AND s.chapter_id = r.chapter_id
  AND s.sort_order = 1;

ALTER TABLE chapter_resources ALTER COLUMN section_id SET NOT NULL;

ALTER TABLE chapter_resources DROP CONSTRAINT IF EXISTS chapter_resources_chapter_id_sort_order_key;

DROP INDEX IF EXISTS uk_chapter_resources_course_chapter_title;

CREATE UNIQUE INDEX uk_chapter_resources_section_sort ON chapter_resources(section_id, sort_order);
CREATE UNIQUE INDEX uk_chapter_resources_section_title ON chapter_resources(course_id, chapter_id, section_id, title);

CREATE INDEX IF NOT EXISTS idx_resources_section_id ON chapter_resources(section_id);

COMMIT;
