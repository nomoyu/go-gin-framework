package nomoyu

import (
	"fmt"
	"github.com/nomoyu/go-gin-framework/internal/router"
	"github.com/nomoyu/go-gin-framework/pkg/config"
)

// InitRemoteConfigIfPresent 如果配置文件定义远程配置中心，则初始化远程配置中心
// 需要在初始化gin框架之后再初始化，因为需要内置路由实现
func initRemoteConfigIfPresent(a *App) {
	if config.Conf.Config.Remote.Addr == "" {
		return
	}
	fmt.Println("[remote config]start init remote config center...")
	// 注册路由
	router.RegisterConfigRoutes(a.engine)

	fmt.Println("[remote config]success init remote config center!")
}
