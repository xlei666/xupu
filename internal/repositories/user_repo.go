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
	ErrUserNotFound     = errors.New("用户不存在")
	ErrUserAlreadyExists = errors.New("用户已存在")
	ErrEmailAlreadyExists = errors.New("邮箱已被注册")
	ErrUsernameAlreadyExists = errors.New("用户名已被使用")
)

// UserRepository 用户仓储
type UserRepository struct {
	db *gorm.DB
}

// NewUserRepository 创建用户仓储
func NewUserRepository() *UserRepository {
	return &UserRepository{
		db: gormdb.Get(),
	}
}

// Create 创建用户
func (r *UserRepository) Create(ctx context.Context, user *models.User) error {
	result := r.db.WithContext(ctx).Create(user)
	if result.Error != nil {
		// 检查是否是唯一约束错误
		if errors.Is(result.Error, gorm.ErrDuplicatedKey) {
			return ErrUserAlreadyExists
		}
		return result.Error
	}
	return nil
}

// GetByID 根据ID获取用户
func (r *UserRepository) GetByID(ctx context.Context, id string) (*models.User, error) {
	var user models.User
	result := r.db.WithContext(ctx).Where("id = ?", id).First(&user)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, result.Error
	}
	return &user, nil
}

// GetByEmail 根据邮箱获取用户
func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User
	result := r.db.WithContext(ctx).Where("email = ?", email).First(&user)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, result.Error
	}
	return &user, nil
}

// GetByUsername 根据用户名获取用户
func (r *UserRepository) GetByUsername(ctx context.Context, username string) (*models.User, error) {
	var user models.User
	result := r.db.WithContext(ctx).Where("username = ?", username).First(&user)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, result.Error
	}
	return &user, nil
}

// Update 更新用户
func (r *UserRepository) Update(ctx context.Context, user *models.User) error {
	result := r.db.WithContext(ctx).Save(user)
	return result.Error
}

// UpdateLastLogin 更新最后登录时间
func (r *UserRepository) UpdateLastLogin(ctx context.Context, userID string) error {
	result := r.db.WithContext(ctx).Model(&models.User{}).
		Where("id = ?", userID).
		Update("last_login_at", gorm.Expr("NOW()"))
	return result.Error
}

// List 列出用户（分页）
func (r *UserRepository) List(ctx context.Context, offset, limit int) ([]models.User, int64, error) {
	var users []models.User
	var total int64

	// 计算总数
	if err := r.db.WithContext(ctx).Model(&models.User{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 获取分页数据
	result := r.db.WithContext(ctx).Offset(offset).Limit(limit).Find(&users)
	if result.Error != nil {
		return nil, 0, result.Error
	}

	return users, total, nil
}

// Delete 删除用户（软删除）
func (r *UserRepository) Delete(ctx context.Context, userID string) error {
	result := r.db.WithContext(ctx).Delete(&models.User{}, "id = ?", userID)
	return result.Error
}

// ExistsByEmail 检查邮箱是否存在
func (r *UserRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	var count int64
	result := r.db.WithContext(ctx).Model(&models.User{}).Where("email = ?", email).Count(&count)
	if result.Error != nil {
		return false, result.Error
	}
	return count > 0, nil
}

// ExistsByUsername 检查用户名是否存在
func (r *UserRepository) ExistsByUsername(ctx context.Context, username string) (bool, error) {
	var count int64
	result := r.db.WithContext(ctx).Model(&models.User{}).Where("username = ?", username).Count(&count)
	if result.Error != nil {
		return false, result.Error
	}
	return count > 0, nil
}
