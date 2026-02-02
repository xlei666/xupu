// Package repositories 数据访问层
package repositories

import (
	"context"
	"errors"

	"github.com/xlei/xupu/internal/models"
	gormdb "github.com/xlei/xupu/pkg/gormdb"
	"gorm.io/gorm"
)

var (
	ErrChapterNotFound      = errors.New("章节不存在")
	ErrChapterAlreadyExists = errors.New("章节已存在")
)

// ChapterRepository 章节仓储
type ChapterRepository struct {
	db *gorm.DB
}

// NewChapterRepository 创建章节仓储
func NewChapterRepository() *ChapterRepository {
	return &ChapterRepository{
		db: gormdb.Get(),
	}
}

// Create 创建章节
func (r *ChapterRepository) Create(ctx context.Context, chapter *models.Chapter) error {
	result := r.db.WithContext(ctx).Create(chapter)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

// GetByID 根据ID获取章节
func (r *ChapterRepository) GetByID(ctx context.Context, id string) (*models.Chapter, error) {
	var chapter models.Chapter
	result := r.db.WithContext(ctx).Where("id = ?", id).First(&chapter)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, ErrChapterNotFound
		}
		return nil, result.Error
	}
	return &chapter, nil
}

// GetByProjectIDAndChapterNum 根据项目ID和章节号获取章节
func (r *ChapterRepository) GetByProjectIDAndChapterNum(ctx context.Context, projectID string, chapterNum int) (*models.Chapter, error) {
	var chapter models.Chapter
	result := r.db.WithContext(ctx).
		Where("project_id = ? AND chapter_num = ?", projectID, chapterNum).
		First(&chapter)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, ErrChapterNotFound
		}
		return nil, result.Error
	}
	return &chapter, nil
}

// ListByProjectID 获取项目的所有章节（按章节号排序）
func (r *ChapterRepository) ListByProjectID(ctx context.Context, projectID string) ([]models.Chapter, error) {
	var chapters []models.Chapter
	result := r.db.WithContext(ctx).
		Where("project_id = ?", projectID).
		Order("chapter_num ASC").
		Find(&chapters)
	if result.Error != nil {
		return nil, result.Error
	}
	return chapters, nil
}

// Update 更新章节
func (r *ChapterRepository) Update(ctx context.Context, chapter *models.Chapter) error {
	result := r.db.WithContext(ctx).Save(chapter)
	return result.Error
}

// UpdateContent 更新章节内容
func (r *ChapterRepository) UpdateContent(ctx context.Context, chapterID string, content string, wordCount int) error {
	result := r.db.WithContext(ctx).Model(&models.Chapter{}).
		Where("id = ?", chapterID).
		Updates(map[string]interface{}{
			"content":    content,
			"word_count": wordCount,
		})
	return result.Error
}

// Delete 删除章节
func (r *ChapterRepository) Delete(ctx context.Context, chapterID string) error {
	result := r.db.WithContext(ctx).Delete(&models.Chapter{}, "id = ?", chapterID)
	return result.Error
}

// DeleteByProjectID 删除项目的所有章节
func (r *ChapterRepository) DeleteByProjectID(ctx context.Context, projectID string) error {
	result := r.db.WithContext(ctx).Delete(&models.Chapter{}, "project_id = ?", projectID)
	return result.Error
}

// CountByProjectID 统计项目的章节数量
func (r *ChapterRepository) CountByProjectID(ctx context.Context, projectID string) (int64, error) {
	var count int64
	result := r.db.WithContext(ctx).Model(&models.Chapter{}).
		Where("project_id = ?", projectID).
		Count(&count)
	if result.Error != nil {
		return 0, result.Error
	}
	return count, nil
}

// GetMaxChapterNum 获取项目的最大章节号
func (r *ChapterRepository) GetMaxChapterNum(ctx context.Context, projectID string) (int, error) {
	var maxChapterNum int
	result := r.db.WithContext(ctx).Model(&models.Chapter{}).
		Where("project_id = ?", projectID).
		Select("COALESCE(MAX(chapter_num), 0)").
		Scan(&maxChapterNum)
	if result.Error != nil {
		return 0, result.Error
	}
	return maxChapterNum, nil
}

// ReorderChapters 重新排序章节（批量更新章节号）
func (r *ChapterRepository) ReorderChapters(ctx context.Context, chapters []models.Chapter) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, chapter := range chapters {
			if err := tx.Save(&chapter).Error; err != nil {
				return err
			}
		}
		return nil
	})
}
