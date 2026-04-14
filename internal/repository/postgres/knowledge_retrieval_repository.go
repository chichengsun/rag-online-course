package postgres

import (
	"context"
	"fmt"
	"strings"

	"gorm.io/gorm"
)

// RetrievalChunk 表示检索到的候选分块。
type RetrievalChunk struct {
	ChunkID      int64
	ResourceID   int64
	ResourceTitle string
	ChunkIndex   int
	Content      string
	Distance     float64
	Score        float64
}

// KnowledgeRetrievalRepository 执行课程范围内向量检索。
type KnowledgeRetrievalRepository struct {
	db *gorm.DB
}

// NewKnowledgeRetrievalRepository 创建检索仓库。
func NewKnowledgeRetrievalRepository(db *gorm.DB) *KnowledgeRetrievalRepository {
	return &KnowledgeRetrievalRepository{db: db}
}

// SearchByCourse 在课程范围内按向量距离召回候选分块（仅已嵌入数据）。
func (r *KnowledgeRetrievalRepository) SearchByCourse(ctx context.Context, teacherID, courseID int64, queryVectorLiteral string, limit int) ([]RetrievalChunk, error) {
	if limit <= 0 {
		limit = 10
	}
	var rows []struct {
		ChunkID      int64   `gorm:"column:chunk_id"`
		ResourceID   int64   `gorm:"column:resource_id"`
		ResourceTitle string  `gorm:"column:resource_title"`
		ChunkIndex   int     `gorm:"column:chunk_index"`
		Content      string  `gorm:"column:content"`
		Distance     float64 `gorm:"column:distance"`
		Score        float64 `gorm:"column:score"`
	}
	err := r.db.WithContext(ctx).Raw(`
		SELECT
		  rec.id AS chunk_id,
		  rec.resource_id AS resource_id,
		  r.title AS resource_title,
		  rec.chunk_index AS chunk_index,
		  rec.content AS content,
		  (rec.embedding <=> ?::vector) AS distance,
		  (1 - (rec.embedding <=> ?::vector)) AS score
		FROM resource_embedding_chunks rec
		INNER JOIN chapter_resources r ON r.id = rec.resource_id
		INNER JOIN course_chapters c ON c.id = r.chapter_id
		INNER JOIN courses co ON co.id = c.course_id
		WHERE co.id = ? AND co.teacher_id = ? AND rec.embedding IS NOT NULL
		ORDER BY rec.embedding <=> ?::vector ASC
		LIMIT ?
	`, queryVectorLiteral, queryVectorLiteral, courseID, teacherID, queryVectorLiteral, limit).Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	out := make([]RetrievalChunk, 0, len(rows))
	for _, it := range rows {
		out = append(out, RetrievalChunk{
			ChunkID:      it.ChunkID,
			ResourceID:   it.ResourceID,
			ResourceTitle: it.ResourceTitle,
			ChunkIndex:   it.ChunkIndex,
			Content:      it.Content,
			Distance:     it.Distance,
			Score:        it.Score,
		})
	}
	return out, nil
}

// SearchByCourseKeyword 在课程范围内按关键词召回候选分块（content/title 模糊匹配）。
func (r *KnowledgeRetrievalRepository) SearchByCourseKeyword(ctx context.Context, teacherID, courseID int64, query string, limit int) ([]RetrievalChunk, error) {
	if limit <= 0 {
		limit = 10
	}
	normalized := strings.TrimSpace(query)
	if normalized == "" {
		return []RetrievalChunk{}, nil
	}
	var rows []struct {
		ChunkID       int64   `gorm:"column:chunk_id"`
		ResourceID    int64   `gorm:"column:resource_id"`
		ResourceTitle string  `gorm:"column:resource_title"`
		ChunkIndex    int     `gorm:"column:chunk_index"`
		Content       string  `gorm:"column:content"`
		Distance      float64 `gorm:"column:distance"`
		Score         float64 `gorm:"column:score"`
	}
	err := r.db.WithContext(ctx).Raw(`
		SELECT
		  rec.id AS chunk_id,
		  rec.resource_id AS resource_id,
		  r.title AS resource_title,
		  rec.chunk_index AS chunk_index,
		  rec.content AS content,
		  (
		    CASE
		      WHEN POSITION(lower(?) IN lower(rec.content)) > 0 THEN 2
		      WHEN POSITION(lower(?) IN lower(r.title)) > 0 THEN 3
		      ELSE 0
		    END
		  )::float / 3.0 AS score,
		  CASE
		    WHEN POSITION(lower(?) IN lower(rec.content)) > 0 THEN POSITION(lower(?) IN lower(rec.content))::float / 10000.0
		    WHEN POSITION(lower(?) IN lower(r.title)) > 0 THEN POSITION(lower(?) IN lower(r.title))::float / 10000.0 + 0.05
		    ELSE 0.9
		  END AS distance
		FROM resource_embedding_chunks rec
		INNER JOIN chapter_resources r ON r.id = rec.resource_id
		INNER JOIN course_chapters c ON c.id = r.chapter_id
		INNER JOIN courses co ON co.id = c.course_id
		WHERE co.id = ? AND co.teacher_id = ? AND (
		  lower(rec.content) LIKE '%' || replace(lower(?), ' ', '%') || '%'
		  OR lower(r.title) LIKE '%' || replace(lower(?), ' ', '%') || '%'
		)
		ORDER BY distance ASC, rec.id DESC
		LIMIT ?
	`, normalized, normalized, normalized, normalized, normalized, normalized, courseID, teacherID, normalized, normalized, limit).Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	out := make([]RetrievalChunk, 0, len(rows))
	for _, it := range rows {
		out = append(out, RetrievalChunk{
			ChunkID:       it.ChunkID,
			ResourceID:    it.ResourceID,
			ResourceTitle: it.ResourceTitle,
			ChunkIndex:    it.ChunkIndex,
			Content:       it.Content,
			Distance:      it.Distance,
			Score:         it.Score,
		})
	}
	return out, nil
}

// SearchByCourseKeywordTokens 在课程范围内按关键词 token 召回候选分块。
func (r *KnowledgeRetrievalRepository) SearchByCourseKeywordTokens(ctx context.Context, teacherID, courseID int64, tokens []string, limit int) ([]RetrievalChunk, error) {
	if limit <= 0 {
		limit = 10
	}
	normalized := make([]string, 0, len(tokens))
	for _, t := range tokens {
		s := strings.TrimSpace(strings.ToLower(t))
		if s != "" {
			normalized = append(normalized, s)
		}
	}
	if len(normalized) == 0 {
		return []RetrievalChunk{}, nil
	}
	scoreParts := make([]string, 0, len(normalized)*2)
	whereParts := make([]string, 0, len(normalized)*2)
	args := make([]any, 0, len(normalized)*6+3)
	for range normalized {
		scoreParts = append(scoreParts, "CASE WHEN lower(rec.content) LIKE ? THEN 2 ELSE 0 END")
		scoreParts = append(scoreParts, "CASE WHEN lower(r.title) LIKE ? THEN 3 ELSE 0 END")
		whereParts = append(whereParts, "lower(rec.content) LIKE ?")
		whereParts = append(whereParts, "lower(r.title) LIKE ?")
	}
	for _, t := range normalized {
		like := "%" + t + "%"
		args = append(args, like, like)
	}
	args = append(args, courseID, teacherID)
	for _, t := range normalized {
		like := "%" + t + "%"
		args = append(args, like, like)
	}
	maxScore := float64(len(normalized) * 5)
	args = append(args, maxScore, maxScore)
	args = append(args, limit)
	scoreExpr := fmt.Sprintf("((%s)::float)", strings.Join(scoreParts, " + "))
	sql := fmt.Sprintf(`
		WITH ranked AS (
		  SELECT
			rec.id AS chunk_id,
			rec.resource_id AS resource_id,
			r.title AS resource_title,
			rec.chunk_index AS chunk_index,
			rec.content AS content,
			%s AS raw_score
		  FROM resource_embedding_chunks rec
		  INNER JOIN chapter_resources r ON r.id = rec.resource_id
		  INNER JOIN course_chapters c ON c.id = r.chapter_id
		  INNER JOIN courses co ON co.id = c.course_id
		  WHERE co.id = ? AND co.teacher_id = ? AND (%s)
		)
		SELECT
		  chunk_id,
		  resource_id,
		  resource_title,
		  chunk_index,
		  content,
		  (raw_score / ?::float) AS score,
		  (1 - (raw_score / ?::float)) AS distance
		FROM ranked
		ORDER BY distance ASC, chunk_id DESC
		LIMIT ?
	`, scoreExpr, strings.Join(whereParts, " OR "))
	var rows []struct {
		ChunkID       int64   `gorm:"column:chunk_id"`
		ResourceID    int64   `gorm:"column:resource_id"`
		ResourceTitle string  `gorm:"column:resource_title"`
		ChunkIndex    int     `gorm:"column:chunk_index"`
		Content       string  `gorm:"column:content"`
		Distance      float64 `gorm:"column:distance"`
		Score         float64 `gorm:"column:score"`
	}
	if err := r.db.WithContext(ctx).Raw(sql, args...).Scan(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]RetrievalChunk, 0, len(rows))
	for _, it := range rows {
		out = append(out, RetrievalChunk{
			ChunkID:       it.ChunkID,
			ResourceID:    it.ResourceID,
			ResourceTitle: it.ResourceTitle,
			ChunkIndex:    it.ChunkIndex,
			Content:       it.Content,
			Distance:      it.Distance,
			Score:         it.Score,
		})
	}
	return out, nil
}
