package db

import (
	"database/sql"
	"fmt"
	"net/url"
	"strings"

	_ "github.com/go-sql-driver/mysql" // 原生驱动用于建库
	_ "github.com/lib/pq"              // 原生驱动用于建库
)

// ensureDatabaseExists 在 gorm.Open 前调用，库不存在时自动创建
func ensureDatabaseExists(opt Option) error {
	switch strings.ToLower(opt.Dialect) {
	case "mysql":
		dbname, serverDSN, err := splitMySQLDSN(opt.DSN)
		if err != nil {
			return err
		}
		if dbname == "" {
			// 没有库名就不处理
			return nil
		}
		// 连接到“无库名”的 DSN
		conn, err := sql.Open("mysql", serverDSN)
		if err != nil {
			return fmt.Errorf("mysql 建库时连接失败: %w", err)
		}
		defer conn.Close()

		// 创建库（utf8mb4）
		_, err = conn.Exec(fmt.Sprintf(
			"CREATE DATABASE IF NOT EXISTS `%s` DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci",
			dbname,
		))
		if err != nil {
			return fmt.Errorf("mysql 创建数据库失败: %w", err)
		}
		return nil

	case "postgres", "postgresql":
		dbname, baseConnStr, err := splitPostgresDSN(opt.DSN)
		if err != nil {
			return err
		}
		if dbname == "" {
			return nil
		}
		// 连接到 postgres 系统库
		conn, err := sql.Open("postgres", baseConnStr+" dbname=postgres")
		if err != nil {
			return fmt.Errorf("postgres 建库时连接失败: %w", err)
		}
		defer conn.Close()

		var exists bool
		row := conn.QueryRow("SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname = $1)", dbname)
		if err := row.Scan(&exists); err != nil {
			return fmt.Errorf("postgres 检查数据库存在失败: %w", err)
		}
		if !exists {
			_, err = conn.Exec(fmt.Sprintf("CREATE DATABASE %q WITH ENCODING 'UTF8'", dbname))
			if err != nil {
				return fmt.Errorf("postgres 创建数据库失败: %w", err)
			}
		}
		return nil

	case "sqlite":
		// sqlite 文件即库，无需创建
		return nil
	default:
		return nil
	}
}

// --- helpers ---

// splitMySQLDSN 分离出 dbname 与“无库名”的 DSN
// 形如：user:pass@tcp(127.0.0.1:3306)/dbname?params
func splitMySQLDSN(dsn string) (dbname string, serverDSN string, err error) {
	// 找到最后一个 '@' 后的第一个 '/' 作为库名开始
	at := strings.LastIndex(dsn, "@")
	if at == -1 {
		// 也可能是 unix socket 或其他形式，但尽量兼容
		// 直接找第一个 '/'（可能错误，但尽力而为）
	}
	slash := strings.Index(dsn[at+1:], "/")
	if slash == -1 {
		// 没带库名
		return "", dsn, nil
	}
	slashPos := at + 1 + slash
	rest := dsn[slashPos+1:]

	// dbname 到 '?' 之前
	q := strings.Index(rest, "?")
	if q == -1 {
		dbname = rest
		serverDSN = dsn[:slashPos] + "/" // 保留末尾 /
		return
	}
	dbname = rest[:q]
	serverDSN = dsn[:slashPos+1] + dsn[slashPos+1+len(dbname):] // 去掉 dbname
	// 把多余的 "/" 规范化成只有一个
	if strings.HasPrefix(serverDSN[slashPos:], "//") {
		serverDSN = dsn[:slashPos+1] + dsn[slashPos+2:]
	}
	return
}

// splitPostgresDSN 分离出 dbname 与“无库名”的连接串（key=value 空格分隔或 URL 形式）
// 支持：
//
//	"host=127.0.0.1 port=5432 user=xx password=xx dbname=nomoyu_go sslmode=disable"
//	"postgres://user:pass@host:5432/nomoyu_go?sslmode=disable"
func splitPostgresDSN(dsn string) (dbname string, baseConnStr string, err error) {
	// URL 形式
	if strings.HasPrefix(dsn, "postgres://") || strings.HasPrefix(dsn, "postgresql://") {
		u, e := url.Parse(dsn)
		if e != nil {
			return "", "", e
		}
		dbname = strings.TrimPrefix(u.Path, "/")
		u.Path = "" // 清除 dbname
		baseConnStr = u.String()
		return
	}

	// key=value 形式
	parts := strings.Fields(dsn)
	var items []string
	for _, p := range parts {
		if strings.HasPrefix(p, "dbname=") {
			dbname = strings.TrimPrefix(p, "dbname=")
			continue
		}
		items = append(items, p)
	}
	baseConnStr = strings.Join(items, " ")
	return
}
