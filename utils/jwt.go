package utils

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var JWTSecret = []byte("feedsystem-dev-secret-key")

type Claims struct {
	AccountID int64 `json:"account_id"`
	jwt.RegisteredClaims
}

func GenerateToken(accountID int64) (string, error) {
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
	return token.SignedString(JWTSecret)
}

// 验证并解析 Token
func ParseToken(tokenStr string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(
		tokenStr,
		&Claims{},
		func(token *jwt.Token) (any, error) { return JWTSecret, nil },
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
	)
	if err != nil {
		return nil, err
	}
	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		// token 里的数据结构符合定义的 Claims 且在有效期内、签名合法
		return claims, nil
	}
	return nil, errors.New("无效的 Token")
}
