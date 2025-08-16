// pkg/redisx/redisx.go
package redisx

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
)

var (
	ErrNotInitialized = errors.New("redis not initialized")
	client            redis.UniversalClient
)

type Config struct {
	Mode         string   // "", "single", "cluster"
	Addr         string   // 单点
	Addrs        []string // 集群
	Password     string
	DB           int
	PoolSize     int
	DialTimeout  time.Duration
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

func Init(cfg Config) error {
	if cfg.PoolSize <= 0 {
		cfg.PoolSize = 20
	}
	// 选择 UniversalClient，统一兼容单点/集群
	uo := &redis.UniversalOptions{
		Addrs:        addrsOf(cfg),
		DB:           cfg.DB,
		Password:     cfg.Password,
		PoolSize:     cfg.PoolSize,
		DialTimeout:  durOr(cfg.DialTimeout, 2*time.Second),
		ReadTimeout:  durOr(cfg.ReadTimeout, 1*time.Second),
		WriteTimeout: durOr(cfg.WriteTimeout, 1*time.Second),
	}
	client = redis.NewUniversalClient(uo)

	// 连接测试
	if err := client.Ping(context.Background()).Err(); err != nil {
		return err
	}
	return nil
}

func addrsOf(c Config) []string {
	if c.Mode == "cluster" || len(c.Addrs) > 0 {
		if len(c.Addrs) > 0 {
			return c.Addrs
		}
	}
	// single
	if c.Addr != "" {
		return []string{c.Addr}
	}
	return nil
}

func durOr(v, def time.Duration) time.Duration {
	if v <= 0 {
		return def
	}
	return v
}

func Close() error {
	if client == nil {
		return nil
	}
	return client.Close()
}

func Client() (redis.UniversalClient, error) {
	if client == nil {
		return nil, ErrNotInitialized
	}
	return client, nil
}

// ============== 便捷方法（够用即可，建议优先使用 Client() 调原生API） ==============

func Set(ctx context.Context, key string, val any, ttl time.Duration) error {
	c, err := Client()
	if err != nil {
		return err
	}
	return c.Set(ctx, key, val, ttl).Err()
}

func GetString(ctx context.Context, key string) (string, error) {
	c, err := Client()
	if err != nil {
		return "", err
	}
	return c.Get(ctx, key).Result()
}

func Del(ctx context.Context, keys ...string) (int64, error) {
	c, err := Client()
	if err != nil {
		return 0, err
	}
	return c.Del(ctx, keys...).Result()
}

func Exists(ctx context.Context, keys ...string) (int64, error) {
	c, err := Client()
	if err != nil {
		return 0, err
	}
	return c.Exists(ctx, keys...).Result()
}

func IncrBy(ctx context.Context, key string, n int64) (int64, error) {
	c, err := Client()
	if err != nil {
		return 0, err
	}
	return c.IncrBy(ctx, key, n).Result()
}

func TTL(ctx context.Context, key string) (time.Duration, error) {
	c, err := Client()
	if err != nil {
		return 0, err
	}
	return c.TTL(ctx, key).Result()
}

func HSet(ctx context.Context, key string, values ...any) error {
	c, err := Client()
	if err != nil {
		return err
	}
	return c.HSet(ctx, key, values...).Err()
}

func HGet(ctx context.Context, key, field string) (string, error) {
	c, err := Client()
	if err != nil {
		return "", err
	}
	return c.HGet(ctx, key, field).Result()
}

// SetJSON JSON 存取（常用）
func SetJSON(ctx context.Context, key string, v any, ttl time.Duration) error {
	b, err := json.Marshal(v)
	if err != nil {
		return err
	}
	return Set(ctx, key, string(b), ttl)
}

func GetJSON(ctx context.Context, key string, out any) error {
	s, err := GetString(ctx, key)
	if err != nil {
		return err
	}
	return json.Unmarshal([]byte(s), out)
}

// RDB 返回已初始化的原生 redis.UniversalClient（未初始化将 panic）
func RDB() redis.UniversalClient {
	if client == nil {
		panic(ErrNotInitialized)
	}
	return client
}

// MustClient 与 RDB 等价：返回已初始化的客户端（未初始化将 panic）
func MustClient() redis.UniversalClient {
	return RDB()
}

// TryClient 返回客户端和是否已初始化的布尔值（不 panic 的尝试式获取）
func TryClient() (redis.UniversalClient, bool) {
	if client == nil {
		return nil, false
	}
	return client, true
}
