package config

import (
	"log"
	"os"
	"time"

	"github.com/spf13/viper"
)

type AppConfig struct {
	App      App           `mapstructure:"app"`
	Server   Server        `mapstructure:"server"`
	Database Database      `mapstructure:"database"`
	Log      Log           `mapstructure:"log"`
	Swagger  SwaggerConfig `mapstructure:"swagger"`
	Auth     AuthConfig    `mapstructure:"auth"`
	Config   ConfigCenter  `mapstructure:"config"`
	Redis    RedisConfig   `mapstructure:"redis"`
	CORS     CORS          `mapstructure:"cors"`
}

type App struct {
	Name    string `mapstructure:"name"`
	Env     string `mapstructure:"env"`
	Version string `mapstructure:"version"`
}

type Server struct {
	Host string `mapstructure:"host"`
	Port int    `mapstructure:"port"`
}

type Database struct {
	Dialect     string `mapstructure:"dialect"`
	Host        string `mapstructure:"host"`
	Port        int    `mapstructure:"port"`
	User        string `mapstructure:"user"`
	Password    string `mapstructure:"password"`
	DBName      string `mapstructure:"dbname"`
	AutoMigrate bool   `mapstructure:"autoMigrate"`
}

type RedisConfig struct {
	// 模式：空或 "single" 使用单点；"cluster" 使用集群（需填写 Addrs）
	Mode     string   `mapstructure:"mode"`
	Addr     string   `mapstructure:"addr"`  // 单点：host:port
	Addrs    []string `mapstructure:"addrs"` // 集群：节点列表
	Password string   `mapstructure:"password"`
	DB       int      `mapstructure:"db"`
	PoolSize int      `mapstructure:"pool_size"`
	// 可选超时（字符串形式，viper 能解析 500ms/2s/1m 等）
	DialTimeout  time.Duration `mapstructure:"dial_timeout"`
	ReadTimeout  time.Duration `mapstructure:"read_timeout"`
	WriteTimeout time.Duration `mapstructure:"write_timeout"`
}

type Log struct {
	Level string `mapstructure:"level"`
	Path  string `mapstructure:"path"`
}

type SwaggerConfig struct {
	Enabled bool   `mapstructure:"enabled"`
	Route   string `mapstructure:"route"`
}

type AuthConfig struct {
	Enabled bool   `mapstructure:"enabled"`
	Mode    string `mapstructure:"mode"`
	JWT     struct {
		Secret string `mapstructure:"secret"`
	} `mapstructure:"jwt"`
}

type RemoteConfig struct {
	Addr string `mapstructure:"addr"`
}

type CORS struct {
	Enabled          bool     `mapstructure:"enabled"`
	AllowOrigins     []string `mapstructure:"allow_origins"`
	AllowMethods     []string `mapstructure:"allow_methods"`
	AllowHeaders     []string `mapstructure:"allow_headers"`
	ExposeHeaders    []string `mapstructure:"expose_headers"`
	AllowCredentials bool     `mapstructure:"allow_credentials"`
	MaxAge           int      `mapstructure:"max_age"` // 秒
}

type ConfigCenter struct {
	Remote RemoteConfig `mapstructure:"remote"`
}

var Conf *AppConfig

func InitConfig() {
	env := os.Getenv("GO_ENV")
	if env == "" {
		env = "dev"
	}

	configFile := "config." + env + ".yaml"
	viper.SetConfigFile(configFile)
	viper.SetConfigType("yaml")

	err := viper.ReadInConfig()
	if err != nil {
		log.Fatalf("读取配置文件失败: %v", err)
	}

	var config AppConfig
	if err := viper.Unmarshal(&config); err != nil {
		log.Fatalf("解析配置文件失败: %v", err)
	}

	Conf = &config
	log.Println("config success init...")
}
