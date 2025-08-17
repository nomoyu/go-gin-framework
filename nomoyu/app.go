package nomoyu

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/nomoyu/go-gin-framework/internal/middleware"
	"github.com/nomoyu/go-gin-framework/pkg/config"
	"github.com/nomoyu/go-gin-framework/pkg/response"
	"net/http"
	"time"
)

type App struct {
	engine          *gin.Engine
	modules         []Module
	routes          []RouteGroup
	logOption       *LogOption
	serverOption    *ServerOption
	authOption      *AuthOption
	dbOption        *DBOption
	httpServer      *http.Server
	shutdownTimeout time.Duration
	redisOption     *RedisOption
	shutting        int32
	shutdownHooks   []func(ctx context.Context) error
	corsOption      *CORSOption
}

func Start() *App {
	config.InitConfig()

	app := &App{
		engine:    gin.New(),
		modules:   []Module{},
		routes:    []RouteGroup{},
		logOption: nil,
	}
	printBanner()
	initLogFromConfigIfPresent(app)
	initRemoteConfigIfPresent(app)
	initSwaggerFromConfigIfPresent(app)
	initDBIfPresent(app)
	initRedisFromConfigIfPresent(app)
	initCORS(app)

	app.engine.Use(
		middleware.TraceID(),
		middleware.RequestContext(),
		middleware.RecoveryMiddleware(),
		middleware.RequestLoggerMiddleware(),
	)
	app.engine.NoRoute(func(c *gin.Context) {
		response.NotFound(c, "无法找到您请求的页面")
	})
	initAuthIfConfigured(app)

	return app
}

func (a *App) WithModule(m Module) *App {
	a.modules = append(a.modules, m)
	return a
}

func (a *App) WithRoute(groups ...RouteGroup) *App {
	a.routes = append(a.routes, groups...)
	return a
}
