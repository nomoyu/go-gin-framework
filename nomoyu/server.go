package nomoyu

import (
	"fmt"
	"github.com/nomoyu/go-gin-framework/internal/middleware"
	"github.com/nomoyu/go-gin-framework/pkg/config"
	"github.com/nomoyu/go-gin-framework/pkg/logger"
)

type ServerOption struct {
	Port     string
	FromUser bool
}

func (a *App) Run(addr ...string) {
	// 注册模块
	for _, m := range a.modules {
		m.Register(a.engine)
	}

	// 注册路由
	for _, group := range a.routes {
		g := a.engine.Group(group.prefix)
		// ✅ 如果启用了权限认证模块并且该路由声明了 RequireAuth
		if group.requireAuth && a.authOption != nil {
			g.Use(middleware.AuthMiddleware(a.authOption.Strategy))
		}
		if len(group.middleware) > 0 {
			g.Use(group.middleware...)
		}
		for _, register := range group.routes {
			register(g)
		}
	}

	// 端口优先级：传参 > 配置 > 默认
	port := ":3303"
	if len(addr) > 0 {
		port = addr[0]
	} else if config.Conf.Server.Port != 0 {
		port = fmt.Sprintf(":%d", config.Conf.Server.Port)
	}

	logger.Infof("✅ 配置初始化完成 " + port)
	// 自定义日志格式
	//a.engine.DebugPrintRouteFunc = func(httpMethod, absolutePath, handlerName string, nuHandlers int) {
	//	log.Printf("endpoint %v %v %v %v\n", httpMethod, absolutePath, handlerName, nuHandlers)
	//}

	if err := a.engine.Run(port); err != nil {
		fmt.Println("服务启动失败: " + err.Error())
	}

}
