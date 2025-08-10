package main

import (
	"github.com/nomoyu/go-gin-framework/nomoyu"
)

func main() {
	nomoyu.Start().
		WithSwagger(). // 未来拓展模块
		Run(":8080")
}
