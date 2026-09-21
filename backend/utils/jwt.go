package utils

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var accessTokenSecret []byte

// 保存 JWT 中的账号 ID 和标准有效期字段（过期、签发、生效时间等）
type AccessTokenClaims struct {
	AccountID int64 `json:"account_id"`
	jwt.RegisteredClaims
}

func InitJWT(secret string) error {
	if strings.TrimSpace(secret) == "" {
		return errors.New("JWT 密钥不能为空")
	}
	accessTokenSecret = []byte(secret)
	return nil
}

// 为指定账号生成 Token
func GenerateAccessToken(accountID int64) (string, error) {
	if len(accessTokenSecret) == 0 {
		return "", errors.New("JWT 尚未初始化")
	}

	accessTokenID, err := generateAccessTokenID()
	if err != nil {
		return "", errors.New("生成 Token ID 失败")
	}

	now := time.Now()
	claims := AccessTokenClaims{
		AccountID: accountID,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        accessTokenID,
			ExpiresAt: jwt.NewNumericDate(now.Add(30 * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
		},
	}

	unsignedAccessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return unsignedAccessToken.SignedString(accessTokenSecret)
}

func GenerateRefreshToken() (string, string, error) {
	refreshToken, err := generateRandomToken(32)
	if err != nil {
		return "", "", err
	}

	return refreshToken, HashRefreshToken(refreshToken), nil
}

func HashRefreshToken(refreshToken string) string {
	sum := sha256.Sum256([]byte(refreshToken))
	return hex.EncodeToString(sum[:])
}

// 验证 Token 并解析用户信息
func ParseAccessToken(accessToken string) (*AccessTokenClaims, error) {
	if len(accessTokenSecret) == 0 {
		return nil, errors.New("JWT 尚未初始化")
	}

	parsedAccessToken, err := jwt.ParseWithClaims(
		accessToken,
		&AccessTokenClaims{},
		func(jwtToken *jwt.Token) (any, error) {
			return accessTokenSecret, nil
		},
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
	)
	if err != nil {
		return nil, err
	}

	claims, ok := parsedAccessToken.Claims.(*AccessTokenClaims)
	if !ok || !parsedAccessToken.Valid {
		return nil, errors.New("无效的 Token")
	}
	return claims, nil
}

// 生成 jti
func generateAccessTokenID() (string, error) {
	return generateRandomToken(16)
}

func generateRandomToken(size int) (string, error) {
	data := make([]byte, size)
	if _, err := rand.Read(data); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(data), nil
}
