package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/nomoyu/go-gin-framework/internal/auth"
	"github.com/nomoyu/go-gin-framework/pkg/response"
	"net/http"
)

func AuthMiddleware(strategy auth.AuthStrategy) gin.HandlerFunc {
	return func(c *gin.Context) {
		authInfo, err := strategy.Authenticate(c)
		if err != nil {
			// 渲染 401 HTML 页面
			response.HTML(c, http.StatusUnauthorized, "401.html", map[string]interface{}{
				"Message": err.Error(),
			})
			c.Abort()
			return
		}

		// 写入上下文
		c.Set("AuthInfo", authInfo)
		c.Next()
	}
}
