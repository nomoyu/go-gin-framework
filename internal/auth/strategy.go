package auth

import "github.com/gin-gonic/gin"

// AuthStrategy 定义认证策略接口
type AuthStrategy interface {
	Authenticate(c *gin.Context) (map[string]interface{}, error)
}
