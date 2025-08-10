package db

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
)

var (
	once    sync.Once
	inst    *gorm.DB
	initErr error
)

// Option 支持手动传入（WithDB）或从配置拼装
type Option struct {
	Dialect         string // mysql / postgres / sqlite
	DSN             string // 完整 DSN（如果非空，优先用）
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	LogLevel        logger.LogLevel // gorm 日志级别
}

// Model a basic GoLang struct which includes the following fields: ID, CreatedAt, UpdatedAt, DeletedAt
// It may be embedded into your model or you may build your own model without it
//
//	type User struct {
//	  gorm.Model
//	}
type Model struct {
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

// Init 初始化单例（幂等）
func Init(opt Option) error {
	once.Do(func() {
		// 先确保库存在（MySQL/Postgres 有库名时）
		if err := ensureDatabaseExists(opt); err != nil {
			initErr = err
			return
		}
		var dial gorm.Dialector

		if opt.DSN == "" {
			initErr = errors.New("db: DSN 为空")
			return
		}

		switch opt.Dialect {
		case "mysql":
			dial = mysql.Open(opt.DSN)
		case "postgres":
			dial = postgres.Open(opt.DSN)
		case "sqlite":
			dial = sqlite.Open(opt.DSN)
		default:
			initErr = fmt.Errorf("db: 不支持的 dialect=%s", opt.Dialect)
			return
		}

		if opt.LogLevel == 0 {
			opt.LogLevel = logger.Warn
		}

		gcfg := &gorm.Config{
			Logger:                                   logger.Default.LogMode(opt.LogLevel),
			DisableForeignKeyConstraintWhenMigrating: true,
		}

		var err error
		inst, err = gorm.Open(dial, gcfg)
		if err != nil {
			initErr = fmt.Errorf("db: gorm open 失败: %w", err)
			return
		}

		sqlDB, err := inst.DB()
		if err != nil {
			initErr = fmt.Errorf("db: 获取底层连接失败: %w", err)
			return
		}

		// 连接池参数（给默认值，配置可覆盖）
		if opt.MaxOpenConns <= 0 {
			opt.MaxOpenConns = 50
		}
		if opt.MaxIdleConns <= 0 {
			opt.MaxIdleConns = 10
		}
		if opt.ConnMaxLifetime <= 0 {
			opt.ConnMaxLifetime = time.Hour
		}
		sqlDB.SetMaxOpenConns(opt.MaxOpenConns)
		sqlDB.SetMaxIdleConns(opt.MaxIdleConns)
		sqlDB.SetConnMaxLifetime(opt.ConnMaxLifetime)

		// 探活
		if err = sqlDB.Ping(); err != nil {
			initErr = fmt.Errorf("db: ping 失败: %w", err)
			return
		}
	})

	return initErr
}

// DB 返回已初始化的 *gorm.DB
func DB() *gorm.DB {
	return inst
}

// MustDB 未初始化时直接 panic（可选）
func MustDB() *gorm.DB {
	if inst == nil {
		panic("db: 尚未初始化，检查配置或 WithDB")
	}
	return inst
}
