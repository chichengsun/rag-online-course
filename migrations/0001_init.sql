DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'user_role') THEN
        CREATE TYPE user_role AS ENUM ('student', 'teacher');
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'course_status') THEN
        CREATE TYPE course_status AS ENUM ('draft', 'published', 'archived');
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'resource_type') THEN
        CREATE TYPE resource_type AS ENUM ('ppt', 'pdf', 'txt', 'video', 'doc', 'docx', 'audio');
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'enrollment_status') THEN
        CREATE TYPE enrollment_status AS ENUM ('active', 'dropped', 'completed');
    END IF;
END $$;

CREATE TABLE IF NOT EXISTS users (
    id BIGSERIAL PRIMARY KEY,
    email VARCHAR(255) NOT NULL,
    username VARCHAR(64) NOT NULL,
    name VARCHAR(120) NOT NULL,
    password_hash TEXT NOT NULL,
    role user_role NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE UNIQUE INDEX IF NOT EXISTS uk_users_username ON users(username);

CREATE TABLE IF NOT EXISTS courses (
    id BIGSERIAL PRIMARY KEY,
    teacher_id BIGINT NOT NULL,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    status course_status NOT NULL DEFAULT 'draft',
    cover_image_url TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_courses_teacher_id ON courses(teacher_id);
CREATE INDEX IF NOT EXISTS idx_courses_status ON courses(status);
CREATE UNIQUE INDEX IF NOT EXISTS uk_courses_teacher_title ON courses(teacher_id, title);

CREATE TABLE IF NOT EXISTS course_chapters (
    id BIGSERIAL PRIMARY KEY,
    course_id BIGINT NOT NULL,
    title VARCHAR(255) NOT NULL,
    sort_order INT NOT NULL CHECK (sort_order > 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (course_id, sort_order)
);
CREATE INDEX IF NOT EXISTS idx_chapters_course_id ON course_chapters(course_id);
CREATE UNIQUE INDEX IF NOT EXISTS uk_course_chapters_course_title ON course_chapters(course_id, title);

CREATE TABLE IF NOT EXISTS chapter_resources (
    id BIGSERIAL PRIMARY KEY,
    chapter_id BIGINT NOT NULL,
    course_id BIGINT NOT NULL,
    title VARCHAR(255) NOT NULL,
    resource_type resource_type NOT NULL,
    sort_order INT NOT NULL CHECK (sort_order > 0),
    object_key TEXT NOT NULL UNIQUE,
    object_url TEXT NOT NULL,
    mime_type VARCHAR(120) NOT NULL,
    size_bytes BIGINT NOT NULL CHECK (size_bytes >= 0),
    duration_seconds INT NOT NULL DEFAULT 0 CHECK (duration_seconds >= 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (chapter_id, sort_order)
);
CREATE INDEX IF NOT EXISTS idx_resources_chapter_id ON chapter_resources(chapter_id);
CREATE INDEX IF NOT EXISTS idx_resources_course_id ON chapter_resources(course_id);
CREATE INDEX IF NOT EXISTS idx_resources_type ON chapter_resources(resource_type);
CREATE UNIQUE INDEX IF NOT EXISTS uk_chapter_resources_course_chapter_title ON chapter_resources(course_id, chapter_id, title);

CREATE TABLE IF NOT EXISTS course_enrollments (
    id BIGSERIAL PRIMARY KEY,
    course_id BIGINT NOT NULL,
    student_id BIGINT NOT NULL,
    status enrollment_status NOT NULL DEFAULT 'active',
    enrolled_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (course_id, student_id)
);
CREATE INDEX IF NOT EXISTS idx_enrollments_student_id ON course_enrollments(student_id);

CREATE TABLE IF NOT EXISTS learning_progress (
    id BIGSERIAL PRIMARY KEY,
    course_id BIGINT NOT NULL,
    student_id BIGINT NOT NULL,
    last_resource_id BIGINT,
    completed_resources INT NOT NULL DEFAULT 0 CHECK (completed_resources >= 0),
    total_resources INT NOT NULL DEFAULT 0 CHECK (total_resources >= 0),
    progress_percent NUMERIC(5,2) NOT NULL DEFAULT 0 CHECK (progress_percent >= 0 AND progress_percent <= 100),
    last_learned_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (course_id, student_id)
);

CREATE TABLE IF NOT EXISTS resource_learning_records (
    id BIGSERIAL PRIMARY KEY,
    resource_id BIGINT NOT NULL,
    student_id BIGINT NOT NULL,
    started_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    watched_seconds INT NOT NULL DEFAULT 0 CHECK (watched_seconds >= 0),
    progress_percent NUMERIC(5,2) NOT NULL DEFAULT 0 CHECK (progress_percent >= 0 AND progress_percent <= 100),
    is_completed BOOLEAN NOT NULL DEFAULT FALSE,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (resource_id, student_id)
);
