package nomoyu

import (
	"fmt"
	"github.com/nomoyu/go-gin-framework/pkg/config"
	"time"
)

const (
	clCyan  = "\033[36m"
	clGreen = "\033[32m"
	clBold  = "\033[1m"
	clGray  = "\033[90m"
	clReset = "\033[0m"
)

// ASCII Logo（可按需替换）
const nomoyuLogo = `
 ________   ________  _____ ______   ________      ___    ___ ___  ___                 ________  ________     
|\   ___  \|\   __  \|\   _ \  _   \|\   __  \    |\  \  /  /|\  \|\  \               |\   ____\|\   __  \    
\ \  \\ \  \ \  \|\  \ \  \\\__\ \  \ \  \|\  \   \ \  \/  / | \  \\\  \  ____________\ \  \___|\ \  \|\  \   
 \ \  \\ \  \ \  \\\  \ \  \\|__| \  \ \  \\\  \   \ \    / / \ \  \\\  \|\____________\ \  \  __\ \  \\\  \  
  \ \  \\ \  \ \  \\\  \ \  \    \ \  \ \  \\\  \   \/  /  /   \ \  \\\  \|____________|\ \  \|\  \ \  \\\  \ 
   \ \__\\ \__\ \_______\ \__\    \ \__\ \_______\__/  / /      \ \_______\              \ \_______\ \_______\
    \|__| \|__|\|_______|\|__|     \|__|\|_______|\___/ /        \|_______|               \|_______|\|_______|
                                                 \|___|/                                                      


`

// 生成 Banner 字符串
func bannerString() string {
	app := config.Conf.App
	srv := config.Conf.Server

	name := app.Name
	if name == "" {
		name = "nomoyu-go"
	}
	version := app.Version
	if version == "" {
		version = "dev"
	}
	env := app.Env
	if env == "" {
		env = "dev"
	}
	port := srv.Port

	return fmt.Sprintf(
		"%s%s%s\n%s%s v%s%s  %senv=%s%s  %sport=%d%s  %s%s%s\n",
		clCyan, nomoyuLogo, clReset,
		clBold, name, version, clReset,
		clGray, env, clReset,
		clGray, port, clReset,
		clGray, time.Now().Format("2006-01-02 15:04:05"), clReset,
	)
}

// 打印 Banner（在 init 或启动时调用）
func printBanner() {
	fmt.Print(bannerString())
}
