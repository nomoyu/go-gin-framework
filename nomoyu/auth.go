package nomoyu

import (
	"fmt"
	"github.com/nomoyu/go-gin-framework/internal/auth"
	"github.com/nomoyu/go-gin-framework/internal/middleware"
	"github.com/nomoyu/go-gin-framework/pkg/config"
)

type AuthConfig struct {
	Enabled bool   `mapstructure:"enabled"`
	Mode    string `mapstructure:"mode"`
	JWT     struct {
		Secret string `mapstructure:"secret"`
	} `mapstructure:"jwt"`
}

type AuthOption struct {
	Strategy auth.AuthStrategy
	FromUser bool
}

func (a *App) WithAuth(strategy auth.AuthStrategy) *App {
	a.authOption = &AuthOption{
		Strategy: strategy,
		FromUser: true,
	}
	return a
}

func initAuthIfConfigured(app *App) {
	if app.authOption != nil {
		app.engine.Use(middleware.AuthMiddleware(app.authOption.Strategy))
		return
	}

	conf := config.Conf.Auth
	if conf.Enabled {
		switch conf.Mode {
		case "jwt":
			a := &auth.JWTStrategy{Secret: conf.JWT.Secret}
			app.authOption = &AuthOption{
				Strategy: a,
				FromUser: true,
			}
		default:
			fmt.Println("❌ 不支持的认证模式:", conf.Mode)
		}
		fmt.Println("启动鉴权成功")
	}
}
