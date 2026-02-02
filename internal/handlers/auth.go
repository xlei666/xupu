// Package handlers HTTP处理器
package handlers

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/xlei/xupu/internal/models"
	"github.com/xlei/xupu/internal/services/auth"
)

// AuthHandler 认证处理器
type AuthHandler struct {
	authService *auth.AuthService
}

// NewAuthHandler 创建认证处理器
func NewAuthHandler(jwtSecret string) *AuthHandler {
	return &AuthHandler{
		authService: auth.NewAuthService(jwtSecret),
	}
}

// ErrorResponse 错误响应
type ErrorResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

// SuccessResponse 成功响应
type SuccessResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
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

// ChangePasswordRequest 修改密码请求
type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required,min=6"`
	NewPassword string `json:"new_password" binding:"required,min=6"`
}

// ForgotPasswordRequest 忘记密码请求
type ForgotPasswordRequest struct {
	Email string `json:"email" binding:"required,email"`
}

// ResetPasswordRequest 重置密码请求
type ResetPasswordRequest struct {
	Token    string `json:"token" binding:"required"`
	Password string `json:"password" binding:"required,min=6"`
}

// UserData 用户数据（用于响应）
type UserData struct {
	ID             string `json:"id"`
	Username       string `json:"username"`
	Email          string `json:"email"`
	Phone          string `json:"phone,omitempty"`
	AvatarURL      string `json:"avatar_url,omitempty"`
	Tier           string `json:"tier"`
	TierExpiresAt  string `json:"tier_expires_at,omitempty"`
	Status         string `json:"status"`
	EmailVerified  bool   `json:"email_verified"`
	LastLoginAt    string `json:"last_login_at,omitempty"`
	CreatedAt      string `json:"created_at"`
	UpdatedAt      string `json:"updated_at"`
}

// AuthTokens 认证token
type AuthTokens struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

// LoginResponse 登录响应
type LoginResponse struct {
	User  UserData   `json:"user"`
	Tokens AuthTokens `json:"tokens"`
}

// RegisterResponse 注册响应
type RegisterResponse struct {
	User  UserData   `json:"user"`
	Tokens AuthTokens `json:"tokens"`
}

// Login 用户登录
// @Summary 用户登录
// @Description 用户使用用户名或邮箱和密码登录
// @Tags 认证
// @Accept json
// @Produce json
// @Param body body LoginRequest true "登录信息"
// @Success 200 {object} SuccessResponse{data=LoginResponse}
// @Failure 400 {object} ErrorResponse
// @Router /api/v1/auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Success: false,
			Error:   "无效的请求参数: " + err.Error(),
		})
		return
	}

	// 调用认证服务
	user, accessToken, refreshToken, err := h.authService.Login(
		c.Request.Context(),
		auth.LoginRequest{
			UsernameOrEmail: req.UsernameOrEmail,
			Password:       req.Password,
		},
	)

	if err != nil {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Success: false,
			Error:   "用户名/邮箱或密码错误",
		})
		return
	}

	// 构建响应
	c.JSON(http.StatusOK, SuccessResponse{
		Success: true,
		Data: LoginResponse{
			User: modelToUserData(user),
			Tokens: AuthTokens{
				AccessToken:  accessToken,
				RefreshToken: refreshToken,
			},
		},
	})
}

// Register 用户注册
// @Summary 用户注册
// @Description 新用户注册
// @Tags 认证
// @Accept json
// @Produce json
// @Param body body RegisterRequest true "注册信息"
// @Success 200 {object} SuccessResponse{data=RegisterResponse}
// @Failure 400 {object} ErrorResponse
// @Router /api/v1/auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Success: false,
			Error:   "无效的请求参数: " + err.Error(),
		})
		return
	}

	// 调用认证服务
	user, accessToken, refreshToken, err := h.authService.Register(
		c.Request.Context(),
		auth.RegisterRequest{
			Username: req.Username,
			Email:    req.Email,
			Password: req.Password,
		},
	)

	if err != nil {
		if err == auth.ErrUserExists || err == auth.ErrWeakPassword {
			c.JSON(http.StatusBadRequest, ErrorResponse{
				Success: false,
				Error:   err.Error(),
			})
			return
		}
		// 返回具体的错误信息而不是"注册失败"
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Success: false,
			Error:   "注册失败: " + err.Error(),
		})
		return
	}

	// 构建响应
	c.JSON(http.StatusCreated, SuccessResponse{
		Success: true,
		Data: RegisterResponse{
			User: modelToUserData(user),
			Tokens: AuthTokens{
				AccessToken:  accessToken,
				RefreshToken: refreshToken,
			},
		},
	})
}

// Logout 用户登出
// @Summary 用户登出
// @Description 用户登出
// @Tags 认证
// @Produce json
// @Security Bearer
// @Success 200 {object} SuccessResponse
// @Router /api/v1/auth/logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
	// 在客户端删除token即可，服务端可以选择将token加入黑名单
	c.JSON(http.StatusOK, SuccessResponse{
		Success: true,
	})
}

// RefreshToken 刷新token
// @Summary 刷新token
// @Description 使用refresh token刷新access token
// @Tags 认证
// @Accept json
// @Produce json
// @Param body body map[string]string true "refresh_token"
// @Success 200 {object} SuccessResponse{data=AuthTokens}
// @Failure 401 {object} ErrorResponse
// @Router /api/v1/auth/refresh [post]
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Success: false,
			Error:   "无效的请求参数",
		})
		return
	}

	accessToken, refreshToken, err := h.authService.RefreshToken(c.Request.Context(), req.RefreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Success: false,
			Error:   "无效的token",
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Success: true,
		Data: AuthTokens{
			AccessToken:  accessToken,
			RefreshToken: refreshToken,
		},
	})
}

// GetCurrentUser 获取当前用户信息
// @Summary 获取当前用户
// @Description 获取当前登录用户的信息
// @Tags 用户
// @Produce json
// @Security Bearer
// @Success 200 {object} SuccessResponse{data=UserData}
// @Failure 401 {object} ErrorResponse
// @Router /api/v1/users/me [get]
func (h *AuthHandler) GetCurrentUser(c *gin.Context) {
	// 从中间件获取用户信息
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Success: false,
			Error:   "未认证",
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Success: true,
		Data:    toUserData(user.(*auth.LoginRequest)),
	})
}

// ChangePassword 修改密码
// @Summary 修改密码
// @Description 修改当前用户密码
// @Tags 用户
// @Accept json
// @Produce json
// @Security Bearer
// @Param body body ChangePasswordRequest true "密码信息"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Router /api/v1/users/me/password [put]
func (h *AuthHandler) ChangePassword(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Success: false,
			Error:   "未认证",
		})
		return
	}

	_ = userID // TODO: Implement password change logic

	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Success: false,
			Error:   "无效的请求参数",
		})
		return
	}

	// TODO: 实现修改密码逻辑
	c.JSON(http.StatusOK, SuccessResponse{
		Success: true,
	})
}

// ForgotPassword 忘记密码
// @Summary 忘记密码
// @Description 请求密码重置邮件
// @Tags 认证
// @Accept json
// @Produce json
// @Param body body ForgotPasswordRequest true "邮箱"
// @Success 200 {object} SuccessResponse
// @Router /api/v1/auth/forgot-password [post]
func (h *AuthHandler) ForgotPassword(c *gin.Context) {
	var req ForgotPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Success: false,
			Error:   "无效的请求参数",
		})
		return
	}

	// 生成重置token
	token, err := h.authService.GenerateResetToken(c.Request.Context(), req.Email)
	if err != nil {
		c.JSON(http.StatusOK, SuccessResponse{
			// 即使邮箱不存在也返回成功，避免邮箱枚举攻击
			Success: true,
		})
		return
	}

	// TODO: 发送密码重置邮件
	// 这里应该使用邮件服务发送包含重置链接的邮件
	// 重置链接格式: https://yourapp.com/reset-password?token={token}
	_ = token

	// 记录日志
	// log.Printf("Password reset token generated for %s: %s", req.Email, token)

	c.JSON(http.StatusOK, SuccessResponse{
		Success: true,
		Data: gin.H{
			"message": "如果该邮箱已注册，您将收到密码重置邮件",
		},
	})
}

// ResetPassword 重置密码
// @Summary 重置密码
// @Description 使用token重置密码
// @Tags 认证
// @Accept json
// @Produce json
// @Param body body ResetPasswordRequest true "token和新密码"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Router /api/v1/auth/reset-password [post]
func (h *AuthHandler) ResetPassword(c *gin.Context) {
	var req ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Success: false,
			Error:   "无效的请求参数",
		})
		return
	}

	if err := h.authService.ResetPassword(c.Request.Context(), req.Token, req.Password); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Success: false,
			Error:   "无效的或已过期的重置链接",
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Success: true,
		Data: gin.H{
			"message": "密码重置成功",
		},
	})
}

// toUserData 将User模型转换为UserData（用于响应）
func toUserData(user interface{}) UserData {
	// 这个函数实际上不会被调用，因为Login和Register会返回models.User
	// 但保留这里以备将来使用
	return UserData{}
}

// modelToUserData 将models.User转换为UserData
func modelToUserData(user *models.User) UserData {
	data := UserData{
		ID:            user.ID,
		Username:      user.Username,
		Email:         user.Email,
		Phone:         user.Phone,
		AvatarURL:     user.AvatarURL,
		Tier:          user.Tier,
		Status:        user.Status,
		EmailVerified: user.EmailVerified,
		CreatedAt:     user.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:     user.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	if user.TierExpire != nil {
		data.TierExpiresAt = user.TierExpire.Format("2006-01-02T15:04:05Z07:00")
	}

	if user.LastLoginAt != nil {
		data.LastLoginAt = user.LastLoginAt.Format("2006-01-02T15:04:05Z07:00")
	}

	return data
}

// getJWTSecret 从环境变量获取JWT密钥
func getJWTSecret() string {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		// 开发环境使用默认密钥
		secret = "your-secret-key-change-in-production"
	}
	return secret
}

// AuthMiddleware JWT认证中间件
func (h *AuthHandler) AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从Authorization header获取token
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error": gin.H{
					"code":    "UNAUTHORIZED",
					"message": "缺少认证令牌",
				},
			})
			c.Abort()
			return
		}

		// 解析Bearer token
		if len(authHeader) < 7 || authHeader[:7] != "Bearer " {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error": gin.H{
					"code":    "INVALID_TOKEN_FORMAT",
					"message": "无效的令牌格式",
				},
			})
			c.Abort()
			return
		}

		tokenString := authHeader[7:]

		// 验证token并获取用户信息
		user, err := h.authService.ValidateToken(c.Request.Context(), tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error": gin.H{
					"code":    "INVALID_TOKEN",
					"message": "无效的或已过期的令牌",
				},
			})
			c.Abort()
			return
		}

		// 将用户ID存入上下文
		c.Set("user_id", user.ID)
		c.Set("user", user)

		c.Next()
	}
}

// GetUserID 从上下文获取用户ID的辅助函数
func GetUserID(c *gin.Context) (string, bool) {
	userID, exists := c.Get("user_id")
	if !exists {
		return "", false
	}
	return userID.(string), true
}
