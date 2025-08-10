package nomoyu

import (
	"fmt"
	"github.com/nomoyu/go-gin-framework/internal/swagger"
	"github.com/nomoyu/go-gin-framework/pkg/config"
)

func (a *App) WithSwagger() *App {
	conf := config.Conf.Swagger
	a.modules = append(a.modules, swagger.New(conf.Route))
	return a
}

// initSwaggerFromConfigIfPresent 在 Start() 中自动调用（用户未 WithSwagger）
func initSwaggerFromConfigIfPresent(app *App) {
	if app.hasModule("swagger") {
		return // 用户已经通过 WithSwagger 显式开启
	}

	conf := config.Conf.Swagger
	if conf.Enabled {
		app.modules = append(app.modules, swagger.New(conf.Route))
	}
}

func (a *App) hasModule(name string) bool {
	for _, m := range a.modules {
		if fmt.Sprintf("%T", m) == name {
			return true
		}
	}
	return false
}
