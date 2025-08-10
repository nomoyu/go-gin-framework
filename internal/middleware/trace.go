package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/nomoyu/go-gin-framework/pkg/trace"
)

func TraceID() gin.HandlerFunc {
	return func(c *gin.Context) {
		traceID := trace.NewTraceID()
		ctx := trace.WithTraceID(c.Request.Context(), traceID)

		// 替换 request 的 context，便于后续调用
		c.Request = c.Request.WithContext(ctx)

		// 设置在 Header 和 Gin context
		c.Writer.Header().Set("X-Trace-ID", traceID)
		c.Set(trace.TraceIDKey, traceID)

		c.Next()
	}
}
