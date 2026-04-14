CREATE TABLE IF NOT EXISTS course_question_bank_items (
    id BIGSERIAL PRIMARY KEY,
    course_id BIGINT NOT NULL REFERENCES courses(id) ON DELETE CASCADE,
    question_type VARCHAR(32) NOT NULL CHECK (question_type IN (
        'single_choice',
        'multiple_choice',
        'true_false',
        'short_answer',
        'fill_blank'
    )),
    stem TEXT NOT NULL,
    reference_answer TEXT NOT NULL,
    source_file_name VARCHAR(255) NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_question_bank_course
    ON course_question_bank_items(course_id, created_at DESC);
