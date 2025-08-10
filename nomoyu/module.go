package nomoyu

import "github.com/gin-gonic/gin"

// Module 所有模块必须实现该接口，注入路由/中间件/服务等
type Module interface {
	Register(r *gin.Engine)
}
