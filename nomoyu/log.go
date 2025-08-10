package nomoyu

import (
	"github.com/nomoyu/go-gin-framework/pkg/config"
	"github.com/nomoyu/go-gin-framework/pkg/logger"
)

type LogOption struct {
	Path     string
	Level    string
	FromUser bool
}

func (a *App) WithLog(path string, level string) *App {
	a.logOption = &LogOption{
		Path:     path,
		Level:    level,
		FromUser: true,
	}
	logger.InitLoggerWithConfig(path, level)
	return a
}

// ✅ 自动从配置初始化（仅当未设置 WithLog 时调用）
func initLogFromConfigIfPresent(app *App) {
	if app.logOption != nil && app.logOption.FromUser {
		return // 已由用户设置，跳过
	}

	conf := config.Conf.Log
	if conf.Path != "" && conf.Level != "" {
		app.logOption = &LogOption{
			Path:     conf.Path,
			Level:    conf.Level,
			FromUser: false,
		}
		logger.InitLoggerWithConfig(conf.Path, conf.Level)
	}
}
