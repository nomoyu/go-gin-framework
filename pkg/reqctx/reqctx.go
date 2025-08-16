package reqctx

import (
	"context"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type RequestCtx struct {
	// 可取消/超时的“请求生命周期”ctx
	Ctx      context.Context
	Gin      *gin.Context
	TraceID  string
	AuthInfo map[string]any
	ClientIP string
	UA       string

	Values map[string]any // 使用者自存临时键值
}

const ginKey = "nomoyu.reqctx"

var (
	bind sync.Map // goroutine-id -> *RequestCtx
)

// FromGin 从 gin.Context 拿/建一个 RequestCtx，并写回 gin.Context
func FromGin(c *gin.Context) *RequestCtx {
	if v, ok := c.Get(ginKey); ok {
		if rc, ok2 := v.(*RequestCtx); ok2 {
			return rc
		}
	}
	rc := &RequestCtx{
		Ctx:      c.Request.Context(),
		Gin:      c,
		ClientIP: c.ClientIP(),
		UA:       c.Request.UserAgent(),
		Values:   map[string]any{},
	}
	c.Set(ginKey, rc)
	return rc
}

// Bind 绑定到当前 goroutine（请求结束时务必 Unbind）
func Bind(rc *RequestCtx) (unbind func()) {
	bind.Store(curGID(), rc)
	return func() { bind.Delete(curGID()) }
}

// Current 取当前 goroutine 的请求上下文（若不存在，返回 nil）
func Current() *RequestCtx {
	if v, ok := bind.Load(curGID()); ok {
		if rc, ok2 := v.(*RequestCtx); ok2 {
			return rc
		}
	}
	return nil
}

// Ctx 返回当前请求的 context.Context（兜底用 Background）
func Ctx() context.Context {
	if rc := Current(); rc != nil && rc.Ctx != nil {
		return rc.Ctx
	}
	return context.Background()
}

// RedisCtx RedisCtx：基于当前请求 ctx 派生一个有超时的 ctx
func RedisCtx(timeout time.Duration) (context.Context, context.CancelFunc) {
	base := Ctx()
	if timeout > 0 {
		return context.WithTimeout(base, timeout)
	}
	return context.WithCancel(base)
}

// Set 便捷访问、设置自定义值
func Set(key string, val any) {
	if rc := Current(); rc != nil {
		rc.Values[key] = val
	}
}
func Get(key string) (any, bool) {
	if rc := Current(); rc != nil {
		v, ok := rc.Values[key]
		return v, ok
	}
	return nil, false
}
func GetString(key string) (string, bool) {
	if v, ok := Get(key); ok {
		s, ok2 := v.(string)
		return s, ok2
	}
	return "", false
}

// 小工具：获取 goroutine id（与您的 logger 同路数）
func curGID() uint64 {
	var b [64]byte
	n := runtime.Stack(b[:], false)
	s := strings.TrimPrefix(string(b[:n]), "goroutine ")
	i := strings.IndexByte(s, ' ')
	if i < 1 {
		return 0
	}
	id, _ := strconv.ParseUint(s[:i], 10, 64)
	return id
}
