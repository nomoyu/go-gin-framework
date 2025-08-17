package middleware

import (
	"fmt"
	"github.com/nomoyu/go-gin-framework/pkg/errorcode"
	"github.com/nomoyu/go-gin-framework/pkg/logger"
	"github.com/nomoyu/go-gin-framework/pkg/response"
	"net/http"
	"runtime/debug"

	"github.com/gin-gonic/gin"
)

func RecoveryMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if rec := recover(); rec != nil {
				// 打印堆栈错误日志
				logger.Error("panic recovered: " + formatRecover(rec))

				// 返回统一错误响应
				c.AbortWithStatusJSON(http.StatusOK, response.Response{
					Code: errorcode.ServerError.Code,
					Msg:  errorcode.ServerError.Msg,
				})
			}
		}()
		c.Next()
	}
}

func formatRecover(rec any) string {
	return fmt.Sprintf("%v\n%s", rec, debug.Stack())
}
