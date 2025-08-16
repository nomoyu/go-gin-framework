package nomoyu

import (
	"context"
	"github.com/nomoyu/go-gin-framework/pkg/config"
	"github.com/nomoyu/go-gin-framework/pkg/logger"
	"github.com/nomoyu/go-gin-framework/pkg/redisx"
)

type RedisOption struct {
	C        redisx.Config
	FromUser bool
}

// WithRedis 链式：手动配置，优先级最高
func (a *App) WithRedis(cfg redisx.Config) *App {
	a.redisOption = &RedisOption{C: cfg, FromUser: true}
	return a
}

// WithRedisAddr 兼容简写
func (a *App) WithRedisAddr(addr, password string, db int) *App {
	return a.WithRedis(redisx.Config{Addr: addr, Password: password, DB: db})
}

// Start() 里调用的一键初始化（配置优先 & 用户可覆盖）
func initRedisFromConfigIfPresent(app *App) {
	// 若用户已 WithRedis -> 初始化并返回
	logger.Info("start init nomoyu redis...")
	if app.redisOption != nil && app.redisOption.FromUser {
		logger.Info("start init redis with code")
		if err := redisx.Init(app.redisOption.C); err != nil {
			logger.Errorf("init nomoyu redis failed(WithRedis): %v", err)
		} else {
			app.OnShutdown(func(ctx context.Context) error {
				logger.Info("close Redis connect...")
				return redisx.Close()
			})
		}
		return
	}
	logger.Info("start init redis with config")
	// 没有 WithRedis，则尝试从配置读取
	rc := config.Conf.Redis
	// 既支持单点也支持集群；只要给了 addr 或 addrs 就尝试初始化
	if rc.Addr != "" || len(rc.Addrs) > 0 {
		logger.Info("start init redis with config" + rc.Addr)
		cfg := redisx.Config{
			Mode:         rc.Mode,
			Addr:         rc.Addr,
			Addrs:        rc.Addrs,
			Password:     rc.Password,
			DB:           rc.DB,
			PoolSize:     rc.PoolSize,
			DialTimeout:  rc.DialTimeout,
			ReadTimeout:  rc.ReadTimeout,
			WriteTimeout: rc.WriteTimeout,
		}
		if err := redisx.Init(cfg); err != nil {
			logger.Errorf("init nomoyu redis failed: %v", err)
			return
		}
		app.redisOption = &RedisOption{C: cfg, FromUser: false}
		logger.Infof("init nomoyu redis success（%s）", modeLabel(cfg))
		app.OnShutdown(func(ctx context.Context) error {
			logger.Info("close Redis connect...")
			return redisx.Close()
		})
	}
}

func modeLabel(c redisx.Config) string {
	if c.Mode == "cluster" || len(c.Addrs) > 1 {
		return "cluster"
	}
	return "single"
}
