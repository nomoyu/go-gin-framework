package middleware

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/nomoyu/go-gin-framework/internal/logger"
	"github.com/nomoyu/go-gin-framework/pkg/trace"
)

type bodyLogWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w bodyLogWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

func RequestLoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// 原始路径（不带 query）+ 带 query 的显示路径
		rawPath := c.Request.URL.Path
		path := rawPath
		if q := c.Request.URL.RawQuery; q != "" {
			path += "?" + q
			rawPath += "?" + q
		}

		// 列（caller）里要显示的路由名：优先 FullPath，否则用原始 path（处理 404）
		routeOrPath := c.FullPath()
		if routeOrPath == "" {
			routeOrPath = rawPath
		}

		// 绑定 traceId（让后续所有 logger.* 自动显示在第二列）
		traceID := trace.GetTraceID(c.Request.Context())
		unbind := logger.BindTraceForRequest(traceID)
		defer unbind()

		route := c.FullPath()
		if route == "" {
			route = "-"
		}
		reqBody, _ := readAndRestoreBody(c.Request, 2048) // 最多记录 2KB
		// ---- 请求开始（REQ）----
		logger.
			WithRouteColumn(rawPath).Info("<--",
			zap.String("method", c.Request.Method),
			//zap.String("route", route),
			zap.String("body", reqBody),
			//zap.String("path", path),
			//zap.String("ip", c.ClientIP()),
		)

		// 捕获响应体（可截断）
		blw := &bodyLogWriter{ResponseWriter: c.Writer, body: bytes.NewBuffer(nil)}
		c.Writer = blw

		c.Next()

		lat := time.Since(start)
		status := c.Writer.Status()

		resp := blw.body.String()
		if len(resp) > 500 {
			resp = resp[:500] + "...(truncated)"
		}

		auth := c.GetHeader("Authorization")
		if auth != "" {
			auth = redact(auth)
		}

		// ---- 请求结束（RESP）----
		logger.
			WithRouteColumn(rawPath).Info("-->",
			zap.String("method", c.Request.Method),
			//zap.String("path", path),
			//zap.String("route", route),
			zap.Int("status", status),
			zap.Duration("cost", lat),
			//zap.String("auth", auth),
			// 如果不想打印响应体，删除下一行
			zap.String("resp", resp),
		)
	}
}

func redact(s string) string {
	if s == "" {
		return ""
	}
	if strings.HasPrefix(strings.ToLower(s), "bearer ") {
		return "Bearer ****"
	}
	if len(s) <= 8 {
		return "****"
	}
	return s[:4] + "****" + s[len(s)-2:]
}

// 读取并复位请求体；返回日志用的字符串
func readAndRestoreBody(r *http.Request, max int) (string, error) {
	if r.Body == nil {
		return "", nil
	}

	var raw []byte
	var err error

	// 处理 gzip 压缩
	if strings.EqualFold(r.Header.Get("Content-Encoding"), "gzip") {
		zr, e := gzip.NewReader(r.Body)
		if e != nil {
			return "", e
		}
		defer zr.Close()
		raw, err = io.ReadAll(zr)
	} else {
		raw, err = io.ReadAll(r.Body)
	}
	if err != nil {
		return "", err
	}

	// 复位 Body 给后续 handler 使用
	r.Body = io.NopCloser(bytes.NewBuffer(raw))

	// 截断，避免太大
	if max > 0 && len(raw) > max {
		raw = append(raw[:max], []byte("...(truncated)")...)
	}

	// 二进制/表单建议别直接打
	ct := r.Header.Get("Content-Type")
	if strings.HasPrefix(ct, "multipart/") {
		return "[multipart omitted]", nil
	}
	return string(raw), nil

}
