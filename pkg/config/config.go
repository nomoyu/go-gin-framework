package config

import (
	"log"
	"os"

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
	log.Println("✅ 配置初始化完成")
}
