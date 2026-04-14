package postgres

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"
)

// TeacherAIModelRow 映射 teacher_ai_models 表一行。
type TeacherAIModelRow struct {
	ID         int64     `gorm:"column:id"`
	TeacherID  int64     `gorm:"column:teacher_id"`
	Name       string    `gorm:"column:name"`
	ModelType  string    `gorm:"column:model_type"`
	APIBaseURL string    `gorm:"column:api_base_url"`
	ModelID    string    `gorm:"column:model_id"`
	APIKey     string    `gorm:"column:api_key"`
	CreatedAt  time.Time `gorm:"column:created_at"`
	UpdatedAt  time.Time `gorm:"column:updated_at"`
}

// TableName 指定 GORM 表名。
func (TeacherAIModelRow) TableName() string { return "teacher_ai_models" }

// TeacherAIModelRepository 教师侧 AI 模型配置持久化。
type TeacherAIModelRepository struct {
	db *gorm.DB
}

// NewTeacherAIModelRepository 创建仓库。
func NewTeacherAIModelRepository(db *gorm.DB) *TeacherAIModelRepository {
	return &TeacherAIModelRepository{db: db}
}

// ListByTeacher 按教师列出全部模型。
func (r *TeacherAIModelRepository) ListByTeacher(ctx context.Context, teacherID int64) ([]TeacherAIModelRow, error) {
	var rows []TeacherAIModelRow
	err := r.db.WithContext(ctx).Where("teacher_id = ?", teacherID).Order("model_type ASC, id ASC").Find(&rows).Error
	return rows, err
}

// GetByID 按主键与教师加载一行（含 api_key，仅供服务端调用外部 API）。
func (r *TeacherAIModelRepository) GetByID(ctx context.Context, id, teacherID int64) (*TeacherAIModelRow, error) {
	var row TeacherAIModelRow
	err := r.db.WithContext(ctx).Where("id = ? AND teacher_id = ?", id, teacherID).First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &row, nil
}

// Create 插入一行并回填 ID。
func (r *TeacherAIModelRepository) Create(ctx context.Context, row *TeacherAIModelRow) error {
	return r.db.WithContext(ctx).Create(row).Error
}

// Update 更新名称、URL、model 字段；apiKey 非 nil 时同时更新 api_key。
func (r *TeacherAIModelRepository) Update(ctx context.Context, id, teacherID int64, name, apiBaseURL, modelID string, apiKey *string) error {
	updates := map[string]any{
		"name":         name,
		"api_base_url": apiBaseURL,
		"model_id":     modelID,
		"updated_at":   time.Now().UTC(),
	}
	if apiKey != nil {
		updates["api_key"] = *apiKey
	}
	res := r.db.WithContext(ctx).Model(&TeacherAIModelRow{}).Where("id = ? AND teacher_id = ?", id, teacherID).Updates(updates)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

// Delete 删除指定教师的模型配置。
func (r *TeacherAIModelRepository) Delete(ctx context.Context, id, teacherID int64) error {
	res := r.db.WithContext(ctx).Where("id = ? AND teacher_id = ?", id, teacherID).Delete(&TeacherAIModelRow{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}
