package nomoyu

import (
	"github.com/gin-gonic/gin"
)

type RouteGroup struct {
	prefix      string
	routes      []func(rg *gin.RouterGroup)
	middleware  []gin.HandlerFunc
	requireAuth bool
}

// NewGroup 创建分组路由
func NewGroup(prefix string) RouteGroup {
	return RouteGroup{
		prefix:     prefix,
		routes:     []func(rg *gin.RouterGroup){},
		middleware: []gin.HandlerFunc{},
	}
}

// RequireAuth 开启认证（内部使用时统一注册中间件）
func (rg RouteGroup) RequireAuth() RouteGroup {
	rg.requireAuth = true
	return rg
}

// GET 注册 GET 路由
func (rg RouteGroup) GET(path string, handler gin.HandlerFunc) RouteGroup {
	rg.routes = append(rg.routes, func(group *gin.RouterGroup) {
		group.GET(path, handler)
	})
	return rg
}

// POST 注册 POST 路由
func (rg RouteGroup) POST(path string, handler gin.HandlerFunc) RouteGroup {
	rg.routes = append(rg.routes, func(group *gin.RouterGroup) {
		group.POST(path, handler)
	})
	return rg
}

func (rg RouteGroup) PUT(path string, handler gin.HandlerFunc) RouteGroup {
	rg.routes = append(rg.routes, func(group *gin.RouterGroup) {
		group.PUT(path, handler)
	})
	return rg
}

func (rg RouteGroup) DELETE(path string, handler gin.HandlerFunc) RouteGroup {
	rg.routes = append(rg.routes, func(group *gin.RouterGroup) {
		group.DELETE(path, handler)
	})
	return rg
}

// Use 添加中间件
func (rg RouteGroup) Use(m ...gin.HandlerFunc) RouteGroup {
	rg.middleware = append(rg.middleware, m...)
	return rg
}
