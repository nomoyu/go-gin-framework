// nomoyu/cors.go
package nomoyu

import (
	"time"

	"github.com/gin-contrib/cors"
	"github.com/nomoyu/go-gin-framework/internal/middleware"
	"github.com/nomoyu/go-gin-framework/pkg/config"
	"github.com/nomoyu/go-gin-framework/pkg/logger"
)

type CORSOption struct {
	AllowOrigins     []string
	AllowMethods     []string
	AllowHeaders     []string
	ExposeHeaders    []string
	AllowCredentials bool
	MaxAge           time.Duration
	FromUser         bool
}

func (a *App) WithCORS(opt CORSOption) *App {
	opt.FromUser = true
	a.corsOption = &opt
	return a
}

func initCORS(app *App) {
	// 用户显式 WithCORS 优先
	if app.corsOption == nil || !app.corsOption.FromUser {
		conf := config.Conf.CORS
		if conf.Enabled {
			app.corsOption = &CORSOption{
				AllowOrigins:     conf.AllowOrigins,
				AllowMethods:     firstOr(conf.AllowMethods, []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}),
				AllowHeaders:     firstOr(conf.AllowHeaders, []string{"Origin", "Content-Type", "Accept", "Authorization"}),
				ExposeHeaders:    conf.ExposeHeaders,
				AllowCredentials: conf.AllowCredentials,
				MaxAge:           time.Duration(conf.MaxAge) * time.Second,
				FromUser:         false,
			}
		}
	}

	if app.corsOption == nil {
		// 未启用
		app.engine.Use(middleware.DefaultCORS())
		return
	}

	// 构建 gin-contrib/cors.Config
	cfg := cors.Config{
		AllowOrigins:     app.corsOption.AllowOrigins,
		AllowMethods:     app.corsOption.AllowMethods,
		AllowHeaders:     app.corsOption.AllowHeaders,
		ExposeHeaders:    app.corsOption.ExposeHeaders,
		AllowCredentials: app.corsOption.AllowCredentials,
		MaxAge:           app.corsOption.MaxAge,
	}

	// 注册到 engine（建议尽早挂载）
	app.engine.Use(middleware.CORSMiddleware(cfg))
	logger.Infof("CORS enabled: origins=%v, credentials=%v", cfg.AllowOrigins, cfg.AllowCredentials)
}

func firstOr[T any](v []T, def []T) []T {
	if len(v) > 0 {
		return v
	}
	return def
}
