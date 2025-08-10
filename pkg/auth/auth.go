package auth

import (
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"time"
)

// UserInfo 是认证后返回的用户上下文信息
type UserInfo struct {
	ID    string
	Name  string
	Roles []string
}

type Claims struct {
	ID    string   `json:"sub"`
	Name  string   `json:"name"`
	Roles []string `json:"roles"`
	jwt.RegisteredClaims
}

func GenerateJWT(id, name string, roles []string, secret string, expiry time.Duration) (string, error) {
	claims := Claims{
		ID:    id,
		Name:  name,
		Roles: roles,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

// GenerateJWTFromMap 动态生成 JWT，支持自定义 payload 字段
func GenerateJWTFromMap(payload map[string]interface{}, secret string, expiry time.Duration) (string, error) {
	// 设置标准字段
	claims := jwt.MapClaims{}

	// 添加用户自定义字段
	for k, v := range payload {
		claims[k] = v
	}

	// 设置标准的过期时间和签发时间
	claims["exp"] = time.Now().Add(expiry).Unix()
	claims["iat"] = time.Now().Unix()

	// 创建 token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// 签名
	return token.SignedString([]byte(secret))
}

// GetAuthInfo 提取 Gin Context 中的认证信息（建议中间件使用 c.Set("AuthInfo", map[string]interface{})）
func GetAuthInfo(c *gin.Context) map[string]interface{} {
	raw, exists := c.Get("AuthInfo")
	if !exists {
		return nil
	}

	info, ok := raw.(map[string]interface{})
	if !ok {
		return nil
	}

	return info
}
