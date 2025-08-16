package nomoyu

import (
	"context"
	"github.com/gin-gonic/gin"
	"net/http"
	"sync/atomic"
)

// 关机防护中间件：关机中直接 503，避免新请求排队
func (a *App) shutdownGuard() gin.HandlerFunc {
	return func(c *gin.Context) {
		if atomic.LoadInt32(&a.shutting) == 1 {
			c.Header("Connection", "close")
			c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{
				"code":    http.StatusServiceUnavailable,
				"message": "server is shutting down",
			})
			return
		}
		c.Next()
	}
}

// OnShutdown 可选：注册优雅停机钩子（比如关闭 DB、停止轮询器、刷日志）
func (a *App) OnShutdown(fn func(ctx context.Context) error) *App {
	a.shutdownHooks = append(a.shutdownHooks, fn)
	return a
}
