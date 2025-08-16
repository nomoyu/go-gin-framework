package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/nomoyu/go-gin-framework/pkg/reqctx"
	"github.com/nomoyu/go-gin-framework/pkg/trace"
)

func RequestContext() gin.HandlerFunc {
	return func(c *gin.Context) {
		rc := reqctx.FromGin(c)

		// TraceID（你已有 TraceID 中间件给 request.Context() 注入过）
		if tid := trace.GetTraceID(c.Request.Context()); tid != "" {
			rc.TraceID = tid
		}

		// AuthInfo（鉴权中间件通过 c.Set("AuthInfo", map) 注入）
		if v, ok := c.Get("AuthInfo"); ok {
			if m, ok2 := v.(map[string]any); ok2 {
				rc.AuthInfo = m
			}
		}

		// 绑定到当前 goroutine；请求结束自动解绑
		unbind := reqctx.Bind(rc)
		defer unbind()

		c.Next()
	}
}
