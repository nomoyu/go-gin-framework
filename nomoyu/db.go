package nomoyu

import (
	"fmt"
	"github.com/nomoyu/go-gin-framework/pkg/config"
	"github.com/nomoyu/go-gin-framework/pkg/db"
	"github.com/nomoyu/go-gin-framework/pkg/logger"
	"time"
)

type DBOption struct {
	Opt      db.Option
	FromUser bool
}

// WithDB 手动覆盖（优先级高于配置文件）
func (a *App) WithDB(dialect, dsn string, opts ...func(*db.Option)) *App {
	opt := db.Option{
		Dialect:         dialect,
		DSN:             dsn,
		MaxOpenConns:    50,
		MaxIdleConns:    10,
		ConnMaxLifetime: time.Hour,
	}
	for _, f := range opts {
		f(&opt)
	}
	a.dbOption = &DBOption{Opt: opt, FromUser: true}
	return a
}

// 初始化数据库
func initDBIfPresent(a *App) {
	// 1) 用户手动指定
	if a.dbOption != nil && a.dbOption.FromUser {
		if err := db.Init(a.dbOption.Opt); err != nil {
			fmt.Println("❌ 数据库初始化失败(WithDB):", err)
		} else {
			fmt.Println("✅ 数据库连接成功(WithDB)")
		}
		return
	}

	// 2) 配置文件自动
	conf := config.Conf.Database
	if conf.Dialect != "" && (conf.Host != "" || conf.Dialect == "sqlite") {
		dsn, err := db.BuildDSN(conf.Dialect, conf.Host, conf.Port, conf.User, conf.Password, conf.DBName)
		if err != nil {
			fmt.Println("❌ 生成 DSN 失败:", err)
			return
		}
		opt := db.Option{Dialect: conf.Dialect, DSN: dsn}
		if err := db.Init(opt); err != nil {
			fmt.Println("❌ 数据库初始化失败(配置):", err)
		} else {
			logger.Info("✅ 数据库连接成功(配置)")
			if !conf.AutoMigrate {
				return
			}
			err := db.AutoMigrate()
			if err != nil {
				logger.Infof("❌ AutoMigrate 失败: %v", err)
				return
			}
			logger.Info("✅ 数据库迁移成功")
		}
	}
}
