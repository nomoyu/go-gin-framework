package router

import (
	"embed"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/nomoyu/go-gin-framework/pkg/config"
	"gopkg.in/yaml.v3"
	"html/template"
	"net/http"
	"os"
)

func RegisterConfigRoutes(router *gin.Engine) {
	router.GET("/config", renderConfigPage)
	router.GET("/config/json", getConfigJSON)
	router.POST("/config/save", saveConfig)
}

func saveConfig(c *gin.Context) {
	yamlText := c.PostForm("yaml")

	var temp config.AppConfig
	if err := yaml.Unmarshal([]byte(yamlText), &temp); err != nil {
		c.String(http.StatusBadRequest, fmt.Sprintf("❌ 配置文件格式错误：%v", err))
		return
	}

	err := os.WriteFile("config.remote.yaml", []byte(yamlText), 0644)
	if err != nil {
		c.String(http.StatusInternalServerError, fmt.Sprintf("保存失败: %v", err))
		return
	}

	c.String(http.StatusOK, "✅ 配置已保存成功！")
}
func getConfigJSON(context *gin.Context) {

}

//go:embed template/*
var tmplFS embed.FS

func renderConfigPage(c *gin.Context) {
	// 获取缓存配置或默认配置
	var currentConfig *config.AppConfig
	currentConfig = config.Conf

	// 序列化为 YAML 字符串（用于 textarea 显示）
	yamlBytes, err := yaml.Marshal(currentConfig)
	if err != nil {
		c.String(http.StatusInternalServerError, "配置转换失败: %v", err)
		return
	}

	// 加载模板
	tmplBytes, err := tmplFS.ReadFile("template/config.tmpl")
	if err != nil {
		c.String(http.StatusInternalServerError, "模板加载失败: %v", err)
		return
	}

	tmpl, err := template.New("config").Parse(string(tmplBytes))
	if err != nil {
		c.String(http.StatusInternalServerError, "模板解析失败: %v", err)
		return
	}

	// 渲染模板，传入 YAML 字符串
	c.Header("Content-Type", "text/html; charset=utf-8")
	tmpl.Execute(c.Writer, string(yamlBytes))
}
