package db

import (
	"fmt"
)

// BuildDSN 按 dialect 生成 DSN（配置驱动）
func BuildDSN(dialect, host string, port int, user, password, dbname string) (string, error) {
	switch dialect {
	case "mysql":
		// 例: user:pass@tcp(localhost:3306)/dbname?parseTime=true&loc=Local&charset=utf8mb4
		return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true&loc=Local&charset=utf8mb4",
			user, password, host, port, dbname), nil
	case "postgres":
		// 例: host=localhost user=postgres password=xxx dbname=xxx port=5432 sslmode=disable TimeZone=Asia/Shanghai
		return fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=disable TimeZone=Asia/Shanghai",
			host, user, password, dbname, port), nil
	case "sqlite":
		// 例: ./data/my.db
		if dbname == "" {
			dbname = "app.db"
		}
		return dbname, nil
	default:
		return "", fmt.Errorf("不支持的 dialect=%s", dialect)
	}
}
