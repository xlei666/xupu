// Package auth 认证服务
package auth

import (
	"context"
	"errors"

	"github.com/xlei/xupu/internal/models"
	"github.com/xlei/xupu/internal/repositories"
)

var (
	ErrInvalidCredentials = errors.New("用户名/邮箱或密码错误")
	ErrUserNotFound      = errors.New("用户不存在")
	ErrUserExists        = errors.New("用户已存在")
	ErrInvalidToken      = errors.New("无效的token")
	ErrTokenExpired      = errors.New("token已过期")
	ErrWeakPassword      = errors.New("密码强度不足")
)

// AuthService 认证服务
type AuthService struct {
	userRepo      *repositories.UserRepository
	jwtService    *JWTService
	passwordSvc   *PasswordService
}

// NewAuthService 创建认证服务
func NewAuthService(jwtSecret string) *AuthService {
	return &AuthService{
		userRepo:    repositories.NewUserRepository(),
		jwtService:  NewJWTService(jwtSecret),
		passwordSvc: NewPasswordService(),
	}
}

// LoginRequest 登录请求
type LoginRequest struct {
	UsernameOrEmail string `json:"username_or_email" binding:"required"`
	Password       string `json:"password" binding:"required,min=6"`
}

// RegisterRequest 注册请求
type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=2,max=50"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

// Login 用户登录（支持用户名或邮箱）
func (s *AuthService) Login(ctx context.Context, req LoginRequest) (*models.User, string, string, error) {
	// 先尝试通过用户名查找
	user, err := s.userRepo.GetByUsername(ctx, req.UsernameOrEmail)
	if err != nil {
		// 如果用户名查找失败，再尝试邮箱
		if errors.Is(err, repositories.ErrUserNotFound) {
			user, err = s.userRepo.GetByEmail(ctx, req.UsernameOrEmail)
			if err != nil {
				if errors.Is(err, repositories.ErrUserNotFound) {
					return nil, "", "", ErrInvalidCredentials
				}
				return nil, "", "", err
			}
		} else {
			return nil, "", "", err
		}
	}

	// 验证密码
	if !s.passwordSvc.Check(req.Password, user.PasswordHash) {
		return nil, "", "", ErrInvalidCredentials
	}

	// 检查账户状态
	if user.Status != "active" {
		return nil, "", "", errors.New("账户已被禁用")
	}

	// 生成token
	accessToken, refreshToken, err := s.jwtService.GenerateTokenPair(
		user.ID,
		user.Email,
		user.Tier,
	)
	if err != nil {
		return nil, "", "", err
	}

	// 更新最后登录时间
	if err := s.userRepo.UpdateLastLogin(ctx, user.ID); err != nil {
		// 记录日志但不影响登录
		// log.Printf("Failed to update last login time: %v", err)
	}

	return user, accessToken, refreshToken, nil
}

// Register 用户注册
func (s *AuthService) Register(ctx context.Context, req RegisterRequest) (*models.User, string, string, error) {
	// 验证密码强度
	if err := s.passwordSvc.Validate(req.Password); err != nil {
		return nil, "", "", ErrWeakPassword
	}

	// 检查邮箱是否已存在
	exists, err := s.userRepo.ExistsByEmail(ctx, req.Email)
	if err != nil {
		return nil, "", "", err
	}
	if exists {
		return nil, "", "", repositories.ErrEmailAlreadyExists
	}

	// 检查用户名是否已存在
	exists, err = s.userRepo.ExistsByUsername(ctx, req.Username)
	if err != nil {
		return nil, "", "", err
	}
	if exists {
		return nil, "", "", repositories.ErrUsernameAlreadyExists
	}

	// 哈希密码
	hashedPassword, err := s.passwordSvc.Hash(req.Password)
	if err != nil {
		return nil, "", "", err
	}

	// 创建用户
	user := &models.User{
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: hashedPassword,
		Tier:         "free",
		Status:       "active",
		EmailVerified: false,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, "", "", err
	}

	// 生成token
	accessToken, refreshToken, err := s.jwtService.GenerateTokenPair(
		user.ID,
		user.Email,
		user.Tier,
	)
	if err != nil {
		return nil, "", "", err
	}

	return user, accessToken, refreshToken, nil
}

// RefreshToken 刷新token
func (s *AuthService) RefreshToken(ctx context.Context, refreshToken string) (string, string, error) {
	// 验证refresh token
	claims, err := s.jwtService.ValidateToken(refreshToken)
	if err != nil {
		return "", "", ErrInvalidToken
	}

	// 获取用户信息
	userID := claims.UserID
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return "", "", ErrUserNotFound
	}

	// 检查账户状态
	if user.Status != "active" {
		return "", "", errors.New("账户已被禁用")
	}

	// 生成新的token对
	newAccessToken, newRefreshToken, err := s.jwtService.GenerateTokenPair(
		user.ID,
		user.Email,
		user.Tier,
	)
	if err != nil {
		return "", "", err
	}

	return newAccessToken, newRefreshToken, nil
}

// ValidateToken 验证token
func (s *AuthService) ValidateToken(ctx context.Context, token string) (*models.User, error) {
	claims, err := s.jwtService.ValidateToken(token)
	if err != nil {
		return nil, ErrInvalidToken
	}

	userID := claims.UserID
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, ErrUserNotFound
	}

	return user, nil
}

// GetUserByID 根据ID获取用户
func (s *AuthService) GetUserByID(ctx context.Context, userID string) (*models.User, error) {
	return s.userRepo.GetByID(ctx, userID)
}

// GenerateResetToken 生成密码重置token
func (s *AuthService) GenerateResetToken(ctx context.Context, email string) (string, error) {
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return "", ErrUserNotFound
	}

	return s.jwtService.GenerateResetToken(user.ID)
}

// ResetPassword 重置密码
func (s *AuthService) ResetPassword(ctx context.Context, token, newPassword string) error {
	// 验证token
	userID, err := s.jwtService.ValidateResetToken(token)
	if err != nil {
		return err
	}

	// 验证密码强度
	if err := s.passwordSvc.Validate(newPassword); err != nil {
		return ErrWeakPassword
	}

	// 哈希新密码
	hashedPassword, err := s.passwordSvc.Hash(newPassword)
	if err != nil {
		return err
	}

	// 获取用户
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return ErrUserNotFound
	}

	// 更新密码
	user.PasswordHash = hashedPassword
	return s.userRepo.Update(ctx, user)
}

// UpdateProfile 更新用户资料
func (s *AuthService) UpdateProfile(ctx context.Context, userID string, updates map[string]interface{}) error {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return err
	}

	// 更新允许的字段
	if username, ok := updates["username"].(string); ok {
		user.Username = username
	}
	if phone, ok := updates["phone"].(string); ok {
		user.Phone = phone
	}
	if avatarURL, ok := updates["avatar_url"].(string); ok {
		user.AvatarURL = avatarURL
	}

	return s.userRepo.Update(ctx, user)
}

// ChangePassword 修改密码
func (s *AuthService) ChangePassword(ctx context.Context, userID string, oldPassword, newPassword string) error {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return err
	}

	// 验证旧密码
	if !s.passwordSvc.Check(oldPassword, user.PasswordHash) {
		return errors.New("旧密码错误")
	}

	// 验证新密码强度
	if err := s.passwordSvc.Validate(newPassword); err != nil {
		return ErrWeakPassword
	}

	// 哈希新密码
	hashedPassword, err := s.passwordSvc.Hash(newPassword)
	if err != nil {
		return err
	}

	user.PasswordHash = hashedPassword
	return s.userRepo.Update(ctx, user)
}
