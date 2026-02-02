// Package auth 认证服务
package auth

import (
	"errors"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

// PasswordService 密码服务
type PasswordService struct {
	cost int
}

// NewPasswordService 创建密码服务
func NewPasswordService() *PasswordService {
	return &PasswordService{
		cost: 12, // bcrypt cost factor
	}
}

// Hash 哈希密码
func (s *PasswordService) Hash(password string) (string, error) {
	if password == "" {
		return "", errors.New("密码不能为空")
	}

	bytes, err := bcrypt.GenerateFromPassword([]byte(password), s.cost)
	if err != nil {
		return "", fmt.Errorf("密码哈希失败: %w", err)
	}

	return string(bytes), nil
}

// Check 检查密码
func (s *PasswordService) Check(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// Validate 验证密码强度
func (s *PasswordService) Validate(password string) error {
	if len(password) < 6 {
		return errors.New("密码至少需要6个字符")
	}

	if len(password) > 100 {
		return errors.New("密码最多100个字符")
	}

	return nil
}
