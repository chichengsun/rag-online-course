package postgres

import (
	"context"
	"strings"

	"gorm.io/gorm"
)

// QuestionBankRepository 负责课程题库题目的持久化读写。
type QuestionBankRepository struct {
	db *gorm.DB
}

// NewQuestionBankRepository 创建题库仓储。
func NewQuestionBankRepository(db *gorm.DB) *QuestionBankRepository {
	return &QuestionBankRepository{db: db}
}

// ListByCourse 查询教师在指定课程下的全部题目（无分页，仅保留兼容内部或导出场景）。
func (r *QuestionBankRepository) ListByCourse(ctx context.Context, courseID, teacherID int64) ([]map[string]any, error) {
	rows := make([]map[string]any, 0)
	err := r.db.WithContext(ctx).Raw(`
		SELECT
			q.id::text AS id,
			q.course_id::text AS course_id,
			q.question_type,
			q.stem,
			q.reference_answer,
			q.source_file_name,
			q.created_at,
			q.updated_at
		FROM course_question_bank_items q
		INNER JOIN courses c ON c.id = q.course_id
		WHERE q.course_id = $1 AND c.teacher_id = $2
		ORDER BY q.created_at DESC, q.id DESC
	`, courseID, teacherID).Scan(&rows).Error
	return rows, err
}

// CountByCourse 统计课程下题目条数；keyword、questionType 非空时追加筛选条件。
func (r *QuestionBankRepository) CountByCourse(ctx context.Context, courseID, teacherID int64, keyword, questionType string) (int64, error) {
	kw := strings.TrimSpace(keyword)
	qt := strings.TrimSpace(questionType)
	var n int64
	var err error
	switch {
	case kw == "" && qt == "":
		err = r.db.WithContext(ctx).Raw(`
			SELECT COUNT(*)
			FROM course_question_bank_items q
			INNER JOIN courses c ON c.id = q.course_id
			WHERE q.course_id = $1 AND c.teacher_id = $2
		`, courseID, teacherID).Scan(&n).Error
	case kw != "" && qt == "":
		pattern := "%" + kw + "%"
		err = r.db.WithContext(ctx).Raw(`
			SELECT COUNT(*)
			FROM course_question_bank_items q
			INNER JOIN courses c ON c.id = q.course_id
			WHERE q.course_id = $1 AND c.teacher_id = $2
			  AND (q.stem ILIKE $3 OR q.reference_answer ILIKE $3 OR q.question_type ILIKE $3)
		`, courseID, teacherID, pattern).Scan(&n).Error
	case kw == "" && qt != "":
		err = r.db.WithContext(ctx).Raw(`
			SELECT COUNT(*)
			FROM course_question_bank_items q
			INNER JOIN courses c ON c.id = q.course_id
			WHERE q.course_id = $1 AND c.teacher_id = $2 AND q.question_type = $3
		`, courseID, teacherID, qt).Scan(&n).Error
	default:
		pattern := "%" + kw + "%"
		err = r.db.WithContext(ctx).Raw(`
			SELECT COUNT(*)
			FROM course_question_bank_items q
			INNER JOIN courses c ON c.id = q.course_id
			WHERE q.course_id = $1 AND c.teacher_id = $2 AND q.question_type = $3
			  AND (q.stem ILIKE $4 OR q.reference_answer ILIKE $4 OR q.question_type ILIKE $4)
		`, courseID, teacherID, qt, pattern).Scan(&n).Error
	}
	return n, err
}

// ListByCoursePaged 分页查询题目列表，按创建时间倒序。
func (r *QuestionBankRepository) ListByCoursePaged(ctx context.Context, courseID, teacherID int64, keyword, questionType string, limit, offset int) ([]map[string]any, error) {
	rows := make([]map[string]any, 0)
	kw := strings.TrimSpace(keyword)
	qt := strings.TrimSpace(questionType)
	var err error
	switch {
	case kw == "" && qt == "":
		err = r.db.WithContext(ctx).Raw(`
			SELECT
				q.id::text AS id,
				q.course_id::text AS course_id,
				q.question_type,
				q.stem,
				q.reference_answer,
				q.source_file_name,
				q.created_at,
				q.updated_at
			FROM course_question_bank_items q
			INNER JOIN courses c ON c.id = q.course_id
			WHERE q.course_id = $1 AND c.teacher_id = $2
			ORDER BY q.created_at DESC, q.id DESC
			LIMIT $3 OFFSET $4
		`, courseID, teacherID, limit, offset).Scan(&rows).Error
	case kw != "" && qt == "":
		pattern := "%" + kw + "%"
		err = r.db.WithContext(ctx).Raw(`
			SELECT
				q.id::text AS id,
				q.course_id::text AS course_id,
				q.question_type,
				q.stem,
				q.reference_answer,
				q.source_file_name,
				q.created_at,
				q.updated_at
			FROM course_question_bank_items q
			INNER JOIN courses c ON c.id = q.course_id
			WHERE q.course_id = $1 AND c.teacher_id = $2
			  AND (q.stem ILIKE $3 OR q.reference_answer ILIKE $3 OR q.question_type ILIKE $3)
			ORDER BY q.created_at DESC, q.id DESC
			LIMIT $4 OFFSET $5
		`, courseID, teacherID, pattern, limit, offset).Scan(&rows).Error
	case kw == "" && qt != "":
		err = r.db.WithContext(ctx).Raw(`
			SELECT
				q.id::text AS id,
				q.course_id::text AS course_id,
				q.question_type,
				q.stem,
				q.reference_answer,
				q.source_file_name,
				q.created_at,
				q.updated_at
			FROM course_question_bank_items q
			INNER JOIN courses c ON c.id = q.course_id
			WHERE q.course_id = $1 AND c.teacher_id = $2 AND q.question_type = $3
			ORDER BY q.created_at DESC, q.id DESC
			LIMIT $4 OFFSET $5
		`, courseID, teacherID, qt, limit, offset).Scan(&rows).Error
	default:
		pattern := "%" + kw + "%"
		err = r.db.WithContext(ctx).Raw(`
			SELECT
				q.id::text AS id,
				q.course_id::text AS course_id,
				q.question_type,
				q.stem,
				q.reference_answer,
				q.source_file_name,
				q.created_at,
				q.updated_at
			FROM course_question_bank_items q
			INNER JOIN courses c ON c.id = q.course_id
			WHERE q.course_id = $1 AND c.teacher_id = $2 AND q.question_type = $3
			  AND (q.stem ILIKE $4 OR q.reference_answer ILIKE $4 OR q.question_type ILIKE $4)
			ORDER BY q.created_at DESC, q.id DESC
			LIMIT $5 OFFSET $6
		`, courseID, teacherID, qt, pattern, limit, offset).Scan(&rows).Error
	}
	return rows, err
}

// Create 在课程下创建一条题目记录，并校验课程归属。
func (r *QuestionBankRepository) Create(ctx context.Context, courseID, teacherID int64, questionType, stem, referenceAnswer, sourceFileName string) (int64, error) {
	var insertedID int64
	err := r.db.WithContext(ctx).Raw(`
		INSERT INTO course_question_bank_items(course_id, question_type, stem, reference_answer, source_file_name)
		SELECT c.id, $3, $4, $5, $6
		FROM courses c
		WHERE c.id = $1 AND c.teacher_id = $2
		RETURNING id
	`, courseID, teacherID, questionType, stem, referenceAnswer, sourceFileName).Scan(&insertedID).Error
	if err != nil {
		return 0, err
	}
	if insertedID == 0 {
		return 0, ErrNotFound
	}
	return insertedID, nil
}

// Update 更新题目核心字段（类型、题干、参考答案）。
func (r *QuestionBankRepository) Update(ctx context.Context, itemID, teacherID int64, questionType, stem, referenceAnswer string) error {
	res := r.db.WithContext(ctx).Exec(`
		UPDATE course_question_bank_items q
		SET question_type = $1, stem = $2, reference_answer = $3, updated_at = NOW()
		FROM courses c
		WHERE q.id = $4 AND q.course_id = c.id AND c.teacher_id = $5
	`, questionType, stem, referenceAnswer, itemID, teacherID)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

// Delete 删除课程题目。
func (r *QuestionBankRepository) Delete(ctx context.Context, itemID, teacherID int64) error {
	res := r.db.WithContext(ctx).Exec(`
		DELETE FROM course_question_bank_items q
		USING courses c
		WHERE q.id = $1 AND q.course_id = c.id AND c.teacher_id = $2
	`, itemID, teacherID)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}
