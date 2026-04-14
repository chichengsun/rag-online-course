package postgres

import (
	"context"
	"encoding/json"

	"gorm.io/gorm"
)

// EmbeddingChunkRepository 管理 resource_embedding_chunks；向量以 pgvector 文本格式写入。
type EmbeddingChunkRepository struct {
	db *gorm.DB
}

// NewEmbeddingChunkRepository 创建分块仓库。
func NewEmbeddingChunkRepository(db *gorm.DB) *EmbeddingChunkRepository {
	return &EmbeddingChunkRepository{db: db}
}

// ListByCourse 分页列出课程下可解析资源及其分块统计（教师校验）。
func (r *EmbeddingChunkRepository) ListByCourse(ctx context.Context, courseID, teacherID int64, offset, limit int) ([]map[string]any, int64, error) {
	var total int64
	err := r.db.WithContext(ctx).Raw(`
		SELECT COUNT(1) FROM chapter_resources r
		INNER JOIN course_sections cs ON cs.id = r.section_id
		INNER JOIN course_chapters c ON c.id = r.chapter_id
		INNER JOIN courses co ON co.id = c.course_id
		WHERE c.course_id = ? AND co.teacher_id = ?
		  AND r.resource_type::text IN ('pdf','doc','docx','txt','ppt')
	`, courseID, teacherID).Scan(&total).Error
	if err != nil {
		return nil, 0, err
	}
	rows := make([]map[string]any, 0)
	err = r.db.WithContext(ctx).Raw(`
		SELECT
			r.id::text AS id,
			r.chapter_id::text AS chapter_id,
			r.section_id::text AS section_id,
			c.title AS chapter_title,
			cs.title AS section_title,
			r.title,
			r.resource_type::text AS resource_type,
			COALESCE(SUM(LENGTH(rec.content)), 0)::bigint AS total_chunk_chars,
			COUNT(rec.id) FILTER (WHERE rec.id IS NOT NULL)::bigint AS chunk_count,
			COUNT(rec.id) FILTER (WHERE rec.embedding IS NOT NULL)::bigint AS embedded_count
		FROM chapter_resources r
		INNER JOIN course_sections cs ON cs.id = r.section_id
		INNER JOIN course_chapters c ON c.id = r.chapter_id
		INNER JOIN courses co ON co.id = c.course_id
		LEFT JOIN resource_embedding_chunks rec ON rec.resource_id = r.id
		WHERE c.course_id = ? AND co.teacher_id = ?
		  AND r.resource_type::text IN ('pdf','doc','docx','txt','ppt')
		GROUP BY r.id, r.chapter_id, r.section_id, c.title, c.sort_order, cs.title, cs.sort_order, r.title, r.resource_type, r.sort_order
		ORDER BY c.sort_order ASC, cs.sort_order ASC, r.sort_order ASC, r.id ASC
		LIMIT ? OFFSET ?
	`, courseID, teacherID, limit, offset).Scan(&rows).Error
	return rows, total, err
}

// ListChunksForResource 列出资源下全部分块（教师归属校验）。
func (r *EmbeddingChunkRepository) ListChunksForResource(ctx context.Context, resourceID, teacherID int64) ([]map[string]any, error) {
	rows := make([]map[string]any, 0)
	err := r.db.WithContext(ctx).Raw(`
		SELECT
			rec.id::text AS id,
			rec.chunk_index,
			rec.content,
			rec.char_start,
			rec.char_end,
			rec.token_count,
			rec.metadata_json,
			rec.confirmed_at,
			rec.embedded_at,
			CASE WHEN rec.embedding IS NULL THEN NULL ELSE vector_dims(rec.embedding) END AS embedding_dims
		FROM resource_embedding_chunks rec
		INNER JOIN chapter_resources r ON r.id = rec.resource_id
		INNER JOIN course_chapters c ON c.id = r.chapter_id
		INNER JOIN courses co ON co.id = c.course_id
		WHERE rec.resource_id = ? AND co.teacher_id = ?
		ORDER BY rec.chunk_index ASC
	`, resourceID, teacherID).Scan(&rows).Error
	return rows, err
}

// ReplaceDraftChunks 删除该资源下全部分块后按新列表插入（避免与历史已嵌入块的 chunk_index 冲突；重新分块会清空旧向量）。
func (r *EmbeddingChunkRepository) ReplaceDraftChunks(ctx context.Context, resourceID, teacherID int64, chunks []struct {
	Index    int
	Content  string
	CharS    *int
	CharE    *int
	MetaJSON []byte
}) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		res := tx.Exec(`
			DELETE FROM resource_embedding_chunks rec
			USING chapter_resources r
			INNER JOIN course_chapters c ON c.id = r.chapter_id
			INNER JOIN courses co ON co.id = c.course_id
			WHERE rec.resource_id = r.id AND r.id = ? AND co.teacher_id = ?
		`, resourceID, teacherID)
		if res.Error != nil {
			return res.Error
		}
		for _, ch := range chunks {
			meta := ch.MetaJSON
			if len(meta) == 0 {
				meta = []byte("{}")
			}
			var cs, ce any
			if ch.CharS != nil {
				cs = *ch.CharS
			}
			if ch.CharE != nil {
				ce = *ch.CharE
			}
			ins := tx.Exec(`
				INSERT INTO resource_embedding_chunks (resource_id, chunk_index, content, char_start, char_end, metadata_json)
				SELECT ?, ?, ?, ?, ?, ?::jsonb
				FROM chapter_resources r
				INNER JOIN course_chapters c ON c.id = r.chapter_id
				INNER JOIN courses co ON co.id = c.course_id
				WHERE r.id = ? AND co.teacher_id = ?
			`, resourceID, ch.Index, ch.Content, cs, ce, string(meta), resourceID, teacherID)
			if ins.Error != nil {
				return ins.Error
			}
			if ins.RowsAffected != 1 {
				return ErrNotFound
			}
		}
		return nil
	})
}

// ConfirmDraftChunks 将尚未确认且未嵌入的分块标记为已确认。
func (r *EmbeddingChunkRepository) ConfirmDraftChunks(ctx context.Context, resourceID, teacherID int64) (int64, error) {
	res := r.db.WithContext(ctx).Exec(`
		UPDATE resource_embedding_chunks rec
		SET confirmed_at = NOW(), updated_at = NOW()
		FROM chapter_resources r
		INNER JOIN course_chapters c ON c.id = r.chapter_id
		INNER JOIN courses co ON co.id = c.course_id
		WHERE rec.resource_id = r.id AND r.id = ? AND co.teacher_id = ?
		  AND rec.confirmed_at IS NULL AND rec.embedding IS NULL
	`, resourceID, teacherID)
	return res.RowsAffected, res.Error
}

// PendingEmbedChunk 待写入向量的分块行（仅服务端使用）。
type PendingEmbedChunk struct {
	ID      int64  `gorm:"column:id"`
	Content string `gorm:"column:content"`
}

// ListChunksPendingEmbed 返回待嵌入分块 id 与正文。
func (r *EmbeddingChunkRepository) ListChunksPendingEmbed(ctx context.Context, resourceID, teacherID int64) ([]PendingEmbedChunk, error) {
	var rows []PendingEmbedChunk
	err := r.db.WithContext(ctx).Raw(`
		SELECT rec.id, rec.content
		FROM resource_embedding_chunks rec
		INNER JOIN chapter_resources r ON r.id = rec.resource_id
		INNER JOIN course_chapters c ON c.id = r.chapter_id
		INNER JOIN courses co ON co.id = c.course_id
		WHERE rec.resource_id = ? AND co.teacher_id = ?
		  AND rec.confirmed_at IS NOT NULL AND rec.embedding IS NULL
		ORDER BY rec.chunk_index ASC
	`, resourceID, teacherID).Scan(&rows).Error
	return rows, err
}

// UpdateChunkEmbedding 写入向量（pgvector 文本）与 embedded_at。
func (r *EmbeddingChunkRepository) UpdateChunkEmbedding(ctx context.Context, chunkID, teacherID int64, vectorLiteral string) error {
	res := r.db.WithContext(ctx).Exec(`
		UPDATE resource_embedding_chunks rec
		SET embedding = ?::vector, embedded_at = NOW(), updated_at = NOW()
		FROM chapter_resources r
		INNER JOIN course_chapters c ON c.id = r.chapter_id
		INNER JOIN courses co ON co.id = c.course_id
		WHERE rec.id = ? AND rec.resource_id = r.id AND co.teacher_id = ?
	`, vectorLiteral, chunkID, teacherID)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

// ClearAllChunksForResource 删除该资源下全部分块（教师归属校验）；用于再次预览前清空或手动清空。
func (r *EmbeddingChunkRepository) ClearAllChunksForResource(ctx context.Context, resourceID, teacherID int64) error {
	res := r.db.WithContext(ctx).Exec(`
		DELETE FROM resource_embedding_chunks rec
		USING chapter_resources r
		INNER JOIN course_chapters c ON c.id = r.chapter_id
		INNER JOIN courses co ON co.id = c.course_id
		WHERE rec.resource_id = r.id AND r.id = ? AND co.teacher_id = ?
	`, resourceID, teacherID)
	return res.Error
}

// UpdateChunkContent 更新单条分块正文；编辑后清空向量与确认时间，需重新确认与嵌入。
func (r *EmbeddingChunkRepository) UpdateChunkContent(ctx context.Context, chunkID, resourceID, teacherID int64, content string, charS, charE *int) error {
	var cs, ce any
	if charS != nil {
		cs = *charS
	}
	if charE != nil {
		ce = *charE
	}
	res := r.db.WithContext(ctx).Exec(`
		UPDATE resource_embedding_chunks rec
		SET content = ?, char_start = ?, char_end = ?,
		    embedding = NULL, embedded_at = NULL, confirmed_at = NULL,
		    updated_at = NOW()
		FROM chapter_resources r
		INNER JOIN course_chapters c ON c.id = r.chapter_id
		INNER JOIN courses co ON co.id = c.course_id
		WHERE rec.id = ? AND rec.resource_id = r.id AND r.id = ? AND co.teacher_id = ?
	`, content, cs, ce, chunkID, resourceID, teacherID)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

// DeleteChunkAndRenumber 删除单条分块并按 chunk_index 连续重排（教师归属校验）。
func (r *EmbeddingChunkRepository) DeleteChunkAndRenumber(ctx context.Context, chunkID, resourceID, teacherID int64) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		del := tx.Exec(`
			DELETE FROM resource_embedding_chunks rec
			USING chapter_resources r
			INNER JOIN course_chapters c ON c.id = r.chapter_id
			INNER JOIN courses co ON co.id = c.course_id
			WHERE rec.id = ? AND rec.resource_id = r.id AND r.id = ? AND co.teacher_id = ?
		`, chunkID, resourceID, teacherID)
		if del.Error != nil {
			return del.Error
		}
		if del.RowsAffected == 0 {
			return ErrNotFound
		}
		return tx.Exec(`
			WITH scoped AS (
				SELECT rec.id, rec.chunk_index
				FROM resource_embedding_chunks rec
				INNER JOIN chapter_resources r ON r.id = rec.resource_id
				INNER JOIN course_chapters c ON c.id = r.chapter_id
				INNER JOIN courses co ON co.id = c.course_id
				WHERE rec.resource_id = ? AND co.teacher_id = ?
			), ordered AS (
				SELECT id, (ROW_NUMBER() OVER (ORDER BY chunk_index) - 1)::int AS new_idx
				FROM scoped
			)
			UPDATE resource_embedding_chunks rec
			SET chunk_index = ordered.new_idx, updated_at = NOW()
			FROM ordered
			WHERE rec.id = ordered.id AND rec.chunk_index <> ordered.new_idx
		`, resourceID, teacherID).Error
	})
}

// MetaJSONBytes 辅助构造 metadata JSON 字节。
func MetaJSONBytes(meta map[string]any) []byte {
	if meta == nil {
		return []byte("{}")
	}
	b, _ := json.Marshal(meta)
	return b
}
