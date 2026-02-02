// Package middleware 中间件
package middleware

import (
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/xlei/xupu/internal/services/auth"
)

// AuthMiddleware JWT认证中间件
type AuthMiddleware struct {
	jwtService *auth.JWTService
}

// NewAuthMiddleware 创建认证中间件
func NewAuthMiddleware() *AuthMiddleware {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "your-secret-key-change-in-production"
	}

	return &AuthMiddleware{
		jwtService: auth.NewJWTService(secret),
	}
}

// Authenticate 认证中间件
// 验证JWT token并将用户信息存入context
func (m *AuthMiddleware) Authenticate() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从Authorization header获取token
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   "缺少认证token",
			})
			c.Abort()
			return
		}

		// 解析Bearer token
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   "无效的认证格式",
			})
			c.Abort()
			return
		}

		tokenString := parts[1]

		// 验证token
		claims, err := m.jwtService.ValidateToken(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   "无效的token",
			})
			c.Abort()
			return
		}

		// 将用户信息存入context
		c.Set("user_id", claims.UserID)
		c.Set("user_email", claims.Email)
		c.Set("user_tier", claims.Tier)

		c.Next()
	}
}

// OptionalAuth 可选认证中间件
// 如果提供了token则验证，没有提供也不报错
func (m *AuthMiddleware) OptionalAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.Next()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.Next()
			return
		}

		tokenString := parts[1]
		claims, err := m.jwtService.ValidateToken(tokenString)
		if err == nil {
			c.Set("user_id", claims.UserID)
			c.Set("user_email", claims.Email)
			c.Set("user_tier", claims.Tier)
		}

		c.Next()
	}
}

// RequireTier 要求特定用户等级
// tier: "free", "vip", "svip", "admin"
func (m *AuthMiddleware) RequireTier(tier string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 先检查是否已认证
		userTier, exists := c.Get("user_tier")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   "需要登录",
			})
			c.Abort()
			return
		}

		// 检查等级
		userTierStr := userTier.(string)
		if !hasRequiredTier(userTierStr, tier) {
			c.JSON(http.StatusForbidden, gin.H{
				"success": false,
				"error":   "权限不足",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// hasRequiredTier 检查用户等级是否满足要求
func hasRequiredTier(userTier, requiredTier string) bool {
	tierRank := map[string]int{
		"free": 0,
		"vip":  1,
		"svip": 2,
		"admin": 3,
	}

	userRank := tierRank[userTier]
	requiredRank := tierRank[requiredTier]

	return userRank >= requiredRank
}
