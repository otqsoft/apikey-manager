// Package jwtx JWT 签发与校验（golang-jwt/v5，纯 Go 实现）。
package jwtx

import (
	"errors"
	"time"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/golang-jwt/jwt/v5"
)

// Claims 自定义载荷
type Claims struct {
	UserID   uint   `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

func secret() []byte {
	return []byte(g.Cfg().MustGet(nil, "jwt.secret", "dev-secret").String())
}

func expire() time.Duration {
	sec := g.Cfg().MustGet(nil, "jwt.expire", 604800).Int()
	if sec <= 0 {
		sec = 604800
	}
	return time.Duration(sec) * time.Second
}

// Issue 签发 token
func Issue(userID uint, username, role string) (string, error) {
	now := time.Now()
	claims := Claims{
		UserID:   userID,
		Username: username,
		Role:     role,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(expire())),
			Subject:   username,
		},
	}
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return tok.SignedString(secret())
}

// Parse 校验并解析 token
func Parse(tokenStr string) (*Claims, error) {
	claims := &Claims{}
	tok, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("签名算法非法")
		}
		return secret(), nil
	})
	if err != nil || !tok.Valid {
		return nil, errors.New("token 无效或已过期")
	}
	return claims, nil
}
