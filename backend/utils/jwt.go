package utils

import (
	"errors"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var jwtSecret []byte

// 保存 JWT 中的账号 ID 和标准有效期字段（过期、签发、生效时间等）
type Claims struct {
	AccountID int64 `json:"account_id"`
	jwt.RegisteredClaims
}

func InitJWT(secret string) error {
	if strings.TrimSpace(secret) == "" {
		return errors.New("JWT 密钥不能为空")
	}
	jwtSecret = []byte(secret)
	return nil
}

// 为指定账号生成 Token
func GenerateToken(accountID int64) (string, error) {
	if len(jwtSecret) == 0 {
		return "", errors.New("JWT 尚未初始化")
	}

	now := time.Now()
	claims := Claims{
		AccountID: accountID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(2 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

// 验证 Token 并解析用户信息
func ParseToken(tokenStr string) (*Claims, error) {
	if len(jwtSecret) == 0 {
		return nil, errors.New("JWT 尚未初始化")
	}

	token, err := jwt.ParseWithClaims(
		tokenStr,
		&Claims{},
		func(token *jwt.Token) (any, error) {
			return jwtSecret, nil
		},
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
	)
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("无效的 Token")
	}
	return claims, nil
}
