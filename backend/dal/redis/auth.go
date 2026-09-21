package redis

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	goredis "github.com/redis/go-redis/v9"
)

var rotateRefreshTokenScript = goredis.NewScript(`
	local accountID = redis.call("GET", KEYS[1])

	if not accountID or accountID ~= ARGV[1] then
		return 0
	end

	redis.call("DEL", KEYS[1])
	redis.call("PSETEX", KEYS[2], ARGV[2], accountID)
	return 1
`)

var revokeLoginSessionScript = goredis.NewScript(`
	local accountID = redis.call("GET", KEYS[1])

	if accountID and accountID == ARGV[1] then
		redis.call("DEL", KEYS[1])
	end

	redis.call("PSETEX", KEYS[2], ARGV[2], "1")
	return 1
`)

func refreshTokenKey(refreshTokenHash string) string {
	return "auth:refresh:" + refreshTokenHash
}

func accessTokenBlacklistKey(accessTokenID string) string {
	return "auth:blacklist:" + accessTokenID
}

func SaveRefreshToken(ctx context.Context, refreshTokenHash string, accountID int64, ttl time.Duration) error {
	// 例如：
	// Key:auth:refresh:5f8d7e...
	// Value:637850158561755141
	// TTL:7 天

	key := refreshTokenKey(refreshTokenHash)
	value := strconv.FormatInt(accountID, 10)

	if err := rdb.Set(ctx, key, value, ttl).Err(); err != nil {
		return fmt.Errorf("保存 Refresh Token 失败: %w", err)
	}
	return nil
}

func FindRefreshTokenAccountID(ctx context.Context, refreshTokenHash string) (int64, bool, error) {
	if refreshTokenHash == "" {
		return 0, false, fmt.Errorf("Refresh Token 哈希不能为空")
	}

	value, err := rdb.Get(ctx, refreshTokenKey(refreshTokenHash)).Result()
	if err != nil {
		if errors.Is(err, goredis.Nil) {
			return 0, false, nil
		}
		return 0, false, fmt.Errorf("查询 Refresh Token 失败: %w", err)
	}

	accountID, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return 0, false, fmt.Errorf("解析 Refresh Token 对应账号失败: %w", err)
	}

	return accountID, true, nil
}

func RotateRefreshToken(ctx context.Context, oldRefreshTokenHash, newRefreshTokenHash string, accountID int64, ttl time.Duration) (bool, error) {
	if oldRefreshTokenHash == "" || newRefreshTokenHash == "" {
		return false, fmt.Errorf("Refresh Token 哈希不能为空")
	}
	if accountID <= 0 {
		return false, fmt.Errorf("账号 ID 不合法")
	}
	if ttl <= 0 {
		return false, fmt.Errorf("Refresh Token 有效期必须大于 0")
	}

	result, err := rotateRefreshTokenScript.Run(
		ctx,
		rdb,
		[]string{refreshTokenKey(oldRefreshTokenHash), refreshTokenKey(newRefreshTokenHash)},
		strconv.FormatInt(accountID, 10),
		ttl.Milliseconds(),
	).Int()
	if err != nil {
		return false, fmt.Errorf("轮换 Refresh Token 失败: %w", err)
	}

	return result == 1, nil
}

func RevokeLoginSession(ctx context.Context, refreshTokenHash string, accountID int64, accessTokenID string, accessTokenBlacklistTTL time.Duration) error {
	if refreshTokenHash == "" {
		return fmt.Errorf("Refresh Token 哈希不能为空")
	}
	if accountID <= 0 {
		return fmt.Errorf("账号 ID 不合法")
	}
	if accessTokenID == "" {
		return fmt.Errorf("Access Token ID 不能为空")
	}
	if accessTokenBlacklistTTL <= 0 {
		return fmt.Errorf("Token 黑名单有效期必须大于 0")
	}

	if _, err := revokeLoginSessionScript.Run(
		ctx,
		rdb,
		[]string{refreshTokenKey(refreshTokenHash), accessTokenBlacklistKey(accessTokenID)},
		strconv.FormatInt(accountID, 10),
		accessTokenBlacklistTTL.Milliseconds(),
	).Result(); err != nil {
		return fmt.Errorf("撤销登录状态失败: %w", err)
	}

	return nil
}

func IsAccessTokenRevoked(ctx context.Context, accessTokenID string) (bool, error) {
	if accessTokenID == "" {
		return false, fmt.Errorf("Access Token ID 不能为空")
	}

	exists, err := rdb.Exists(ctx, accessTokenBlacklistKey(accessTokenID)).Result()
	if err != nil {
		return false, fmt.Errorf("检查 Token 撤销状态失败: %w", err)
	}

	return exists > 0, nil
}
