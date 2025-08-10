package middleware

import (
	"bytes"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/nomoyu/go-gin-framework/pkg/logger"
)

// 自定义 ResponseWriter 用于捕获响应内容
type bodyLogWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w bodyLogWriter) Write(b []byte) (int, error) {
	w.body.Write(b) // 缓存响应内容
	return w.ResponseWriter.Write(b)
}

func RequestLoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		method := c.Request.Method
		path := c.Request.URL.Path
		logger.Infof("%s %s", method, path)

		// 包装 ResponseWriter 以拦截响应体
		blw := &bodyLogWriter{body: bytes.NewBufferString(""), ResponseWriter: c.Writer}
		c.Writer = blw
		// 执行处理链
		c.Next()
		// 处理后逻辑
		cost := time.Since(start)
		status := c.Writer.Status()

		var respBody string
		if blw.body.Len() > 0 {
			respBody = blw.body.String()
		}

		// 限制输出长度（避免大文件）
		if len(respBody) > 500 {
			respBody = respBody[:500] + "...(truncated)"
		}

		logger.Infof("%s %s -> %d (%s) => Response: %s", method, path, status, cost, respBody)
	}
}
