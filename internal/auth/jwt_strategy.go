package auth

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"strings"
	"time"
)

// CustomClaims 自定义 JWT Claims（嵌套标准 RegisteredClaims）
type CustomClaims struct {
	Name  string   `json:"name"`
	Roles []string `json:"roles"`
	jwt.RegisteredClaims
}

// JWTStrategy 是 JWT 的认证策略
type JWTStrategy struct {
	Secret string
}

// Authenticate 实现 AuthStrategy 接口
func (s *JWTStrategy) Authenticate(c *gin.Context) (map[string]interface{}, error) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
		return nil, errors.New("missing or malformed token")
	}

	tokenStr := strings.TrimPrefix(authHeader, "Bearer ")

	claims, err := parseJWTToMap(tokenStr, s.Secret)
	if err != nil {
		return nil, err
	}
	return claims, nil
}

// parseJWT 解析并验证 JWT token
func parseJWT(tokenString string, secret string) (*CustomClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})
	if err != nil || !token.Valid {
		return nil, errors.New("invalid token")
	}

	claims, ok := token.Claims.(*CustomClaims)
	if !ok {
		return nil, errors.New("invalid claims structure")
	}

	if claims.ExpiresAt != nil && claims.ExpiresAt.Time.Before(time.Now()) {
		return nil, errors.New("token expired")
	}

	return claims, nil
}

func parseJWTToMap(tokenString string, secret string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})

	if err != nil || !token.Valid {
		return nil, errors.New("invalid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("invalid claims structure")
	}

	// 可校验 exp、iat 等字段
	if exp, ok := claims["exp"].(float64); ok && int64(exp) < time.Now().Unix() {
		return nil, errors.New("token expired")
	}

	return claims, nil
}
