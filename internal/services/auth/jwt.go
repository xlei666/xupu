// Package auth 认证服务
package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// JWTClaims JWT声明
type JWTClaims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	Tier   string `json:"tier"`
	jwt.RegisteredClaims
}

// JWTService JWT服务
type JWTService struct {
	secretKey     []byte
	accessExpiry  time.Duration
	refreshExpiry time.Duration
}

// NewJWTService 创建JWT服务
func NewJWTService(secretKey string) *JWTService {
	return &JWTService{
		secretKey:     []byte(secretKey),
		accessExpiry:  24 * time.Hour,  // Access token 24小时
		refreshExpiry: 30 * 24 * time.Hour, // Refresh token 30天
	}
}

// GenerateAccessToken 生成访问token
func (s *JWTService) GenerateAccessToken(userID, email, tier string) (string, error) {
	claims := JWTClaims{
		UserID: userID,
		Email:  email,
		Tier:   tier,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.accessExpiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "xupu-api",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.secretKey)
}

// GenerateRefreshToken 生成刷新token
func (s *JWTService) GenerateRefreshToken(userID, email, tier string) (string, error) {
	claims := JWTClaims{
		UserID: userID,
		Email:  email,
		Tier:   tier,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.refreshExpiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "xupu-api",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.secretKey)
}

// GenerateTokenPair 生成token对
func (s *JWTService) GenerateTokenPair(userID, email, tier string) (accessToken, refreshToken string, err error) {
	accessToken, err = s.GenerateAccessToken(userID, email, tier)
	if err != nil {
		return "", "", err
	}

	refreshToken, err = s.GenerateRefreshToken(userID, email, tier)
	if err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}

// ValidateToken 验证token
func (s *JWTService) ValidateToken(tokenString string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("无效的签名方法")
		}
		return s.secretKey, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("无效的token")
}

// RefreshAccessToken 刷新访问token
func (s *JWTService) RefreshAccessToken(refreshToken string) (string, error) {
	claims, err := s.ValidateToken(refreshToken)
	if err != nil {
		return "", err
	}

	// 生成新的access token
	return s.GenerateAccessToken(claims.UserID, claims.Email, claims.Tier)
}

// GenerateResetToken 生成密码重置token
func (s *JWTService) GenerateResetToken(userID string) (string, error) {
	claims := JWTClaims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)), // 1小时有效
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "xupu-api",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.secretKey)
}

// ValidateResetToken 验证密码重置token
func (s *JWTService) ValidateResetToken(tokenString string) (string, error) {
	claims, err := s.ValidateToken(tokenString)
	if err != nil {
		return "", err
	}

	// 验证UUID格式
	if _, err := uuid.Parse(claims.UserID); err != nil {
		return "", errors.New("无效的用户ID")
	}

	return claims.UserID, nil
}
