package postgres

import (
	"context"
	"encoding/json"

	"gorm.io/gorm"
)

// KnowledgeChatRepository 管理知识库对话会话与消息持久化。
type KnowledgeChatRepository struct {
	db *gorm.DB
}

// NewKnowledgeChatRepository 创建对话仓库。
func NewKnowledgeChatRepository(db *gorm.DB) *KnowledgeChatRepository {
	return &KnowledgeChatRepository{db: db}
}

// CreateSession 创建会话并校验课程归属教师。
func (r *KnowledgeChatRepository) CreateSession(ctx context.Context, teacherID, courseID int64, title string) (int64, error) {
	var out struct {
		ID int64 `gorm:"column:id"`
	}
	err := r.db.WithContext(ctx).Raw(`
		INSERT INTO knowledge_chat_sessions (teacher_id, course_id, title)
		SELECT ?, c.id, ?
		FROM courses c
		WHERE c.id = ? AND c.teacher_id = ?
		RETURNING id
	`, teacherID, title, courseID, teacherID).Scan(&out).Error
	if err != nil {
		return 0, err
	}
	if out.ID == 0 {
		return 0, ErrNotFound
	}
	return out.ID, nil
}

// GetSession 校验会话归属并返回基础信息。
func (r *KnowledgeChatRepository) GetSession(ctx context.Context, sessionID, teacherID int64) (map[string]any, error) {
	row := map[string]any{}
	err := r.db.WithContext(ctx).Raw(`
		SELECT id::text AS id, teacher_id::text AS teacher_id, course_id::text AS course_id, title, created_at, updated_at
		FROM knowledge_chat_sessions
		WHERE id = ? AND teacher_id = ?
	`, sessionID, teacherID).Scan(&row).Error
	if err != nil {
		return nil, err
	}
	if len(row) == 0 {
		return nil, ErrNotFound
	}
	return row, nil
}

// ListSessions 分页列出会话，支持按课程过滤。
func (r *KnowledgeChatRepository) ListSessions(ctx context.Context, teacherID int64, courseID *int64, offset, limit int) ([]map[string]any, int64, error) {
	var total int64
	if courseID != nil {
		if err := r.db.WithContext(ctx).Raw(`
			SELECT COUNT(1) FROM knowledge_chat_sessions
			WHERE teacher_id = ? AND course_id = ?
		`, teacherID, *courseID).Scan(&total).Error; err != nil {
			return nil, 0, err
		}
	} else {
		if err := r.db.WithContext(ctx).Raw(`
			SELECT COUNT(1) FROM knowledge_chat_sessions WHERE teacher_id = ?
		`, teacherID).Scan(&total).Error; err != nil {
			return nil, 0, err
		}
	}
	rows := make([]map[string]any, 0)
	if courseID != nil {
		err := r.db.WithContext(ctx).Raw(`
			SELECT
			  s.id::text AS id,
			  s.course_id::text AS course_id,
			  s.title,
			  s.created_at,
			  s.updated_at,
			  COALESCE(COUNT(m.id), 0)::bigint AS message_count,
			  MAX(m.created_at) AS last_message_at
			FROM knowledge_chat_sessions s
			LEFT JOIN knowledge_chat_messages m ON m.session_id = s.id
			WHERE s.teacher_id = ? AND s.course_id = ?
			GROUP BY s.id
			ORDER BY s.updated_at DESC, s.id DESC
			LIMIT ? OFFSET ?
		`, teacherID, *courseID, limit, offset).Scan(&rows).Error
		return rows, total, err
	}
	err := r.db.WithContext(ctx).Raw(`
		SELECT
		  s.id::text AS id,
		  s.course_id::text AS course_id,
		  s.title,
		  s.created_at,
		  s.updated_at,
		  COALESCE(COUNT(m.id), 0)::bigint AS message_count,
		  MAX(m.created_at) AS last_message_at
		FROM knowledge_chat_sessions s
		LEFT JOIN knowledge_chat_messages m ON m.session_id = s.id
		WHERE s.teacher_id = ?
		GROUP BY s.id
		ORDER BY s.updated_at DESC, s.id DESC
		LIMIT ? OFFSET ?
	`, teacherID, limit, offset).Scan(&rows).Error
	return rows, total, err
}

// TouchSession 更新时间。
func (r *KnowledgeChatRepository) TouchSession(ctx context.Context, sessionID, teacherID int64) error {
	res := r.db.WithContext(ctx).Exec(`
		UPDATE knowledge_chat_sessions
		SET updated_at = NOW()
		WHERE id = ? AND teacher_id = ?
	`, sessionID, teacherID)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

// UpdateSessionTitle 更新会话标题。
func (r *KnowledgeChatRepository) UpdateSessionTitle(ctx context.Context, sessionID, teacherID int64, title string) error {
	res := r.db.WithContext(ctx).Exec(`
		UPDATE knowledge_chat_sessions
		SET title = ?, updated_at = NOW()
		WHERE id = ? AND teacher_id = ?
	`, title, sessionID, teacherID)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

// DeleteSession 删除会话及其消息。
func (r *KnowledgeChatRepository) DeleteSession(ctx context.Context, sessionID, teacherID int64) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		delMsgs := tx.Exec(`
			DELETE FROM knowledge_chat_messages m
			USING knowledge_chat_sessions s
			WHERE m.session_id = s.id AND s.id = ? AND s.teacher_id = ?
		`, sessionID, teacherID)
		if delMsgs.Error != nil {
			return delMsgs.Error
		}
		delSession := tx.Exec(`
			DELETE FROM knowledge_chat_sessions
			WHERE id = ? AND teacher_id = ?
		`, sessionID, teacherID)
		if delSession.Error != nil {
			return delSession.Error
		}
		if delSession.RowsAffected == 0 {
			return ErrNotFound
		}
		return nil
	})
}

// InsertMessage 新增一条消息。
func (r *KnowledgeChatRepository) InsertMessage(
	ctx context.Context,
	sessionID, teacherID int64,
	role, content string,
	references []map[string]any,
	modelSnapshot map[string]any,
) (int64, error) {
	refJSON := []byte("[]")
	if references != nil {
		if b, err := json.Marshal(references); err == nil {
			refJSON = b
		}
	}
	modelJSON := []byte("{}")
	if modelSnapshot != nil {
		if b, err := json.Marshal(modelSnapshot); err == nil {
			modelJSON = b
		}
	}
	var out struct {
		ID int64 `gorm:"column:id"`
	}
	err := r.db.WithContext(ctx).Raw(`
		INSERT INTO knowledge_chat_messages (session_id, role, content, references_json, model_snapshot_json)
		SELECT s.id, ?, ?, ?::jsonb, ?::jsonb
		FROM knowledge_chat_sessions s
		WHERE s.id = ? AND s.teacher_id = ?
		RETURNING id
	`, role, content, string(refJSON), string(modelJSON), sessionID, teacherID).Scan(&out).Error
	if err != nil {
		return 0, err
	}
	if out.ID == 0 {
		return 0, ErrNotFound
	}
	return out.ID, nil
}

// ListMessages 分页列出会话消息（升序）。
func (r *KnowledgeChatRepository) ListMessages(ctx context.Context, sessionID, teacherID int64, offset, limit int) ([]map[string]any, int64, error) {
	var total int64
	err := r.db.WithContext(ctx).Raw(`
		SELECT COUNT(1)
		FROM knowledge_chat_messages m
		INNER JOIN knowledge_chat_sessions s ON s.id = m.session_id
		WHERE m.session_id = ? AND s.teacher_id = ?
	`, sessionID, teacherID).Scan(&total).Error
	if err != nil {
		return nil, 0, err
	}
	rows := make([]map[string]any, 0)
	err = r.db.WithContext(ctx).Raw(`
		SELECT
		  m.id::text AS id,
		  m.session_id::text AS session_id,
		  m.role,
		  m.content,
		  m.references_json,
		  m.model_snapshot_json,
		  m.created_at
		FROM knowledge_chat_messages m
		INNER JOIN knowledge_chat_sessions s ON s.id = m.session_id
		WHERE m.session_id = ? AND s.teacher_id = ?
		ORDER BY m.created_at ASC, m.id ASC
		LIMIT ? OFFSET ?
	`, sessionID, teacherID, limit, offset).Scan(&rows).Error
	return rows, total, err
}

