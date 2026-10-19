// Package auth 封装 JWT 的签发与校验。
package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// ErrInvalidToken 表示 token 缺失、格式错误、签名不匹配或已过期。
var ErrInvalidToken = errors.New("无效的 token")

// Claims 是 JWT 中携带的自定义声明。
type Claims struct {
	UserID   uint   `json:"user_id"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}

// GenerateToken 为指定用户签发一个 HS256 的 JWT。
func GenerateToken(secret []byte, userID uint, username string, ttl time.Duration) (string, error) {
	now := time.Now()
	claims := Claims{
		UserID:   userID,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   fmt.Sprintf("%d", userID),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(secret)
}

// ParseToken 校验签名与有效期，并返回其中的声明。
func ParseToken(secret []byte, tokenString string) (*Claims, error) {
	claims := &Claims{}

	_, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (any, error) {
		// 必须校验签名算法，否则攻击者可把 alg 改成 none 或非对称算法绕过签名
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("非预期的签名算法: %v", token.Header["alg"])
		}
		return secret, nil
	})
	if err != nil {
		return nil, ErrInvalidToken
	}

	return claims, nil
}
