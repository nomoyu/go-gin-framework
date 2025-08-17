package response

import (
	"embed"
	"github.com/gin-gonic/gin"
	"github.com/nomoyu/go-gin-framework/pkg/errorcode"
	"html/template"
	"net/http"
)

const (
	CodeSuccess = 0
	CodeError   = 500
)

// Response 通用响应结构：泛型版本
type Response struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data,omitempty"` // 空值时可省略
}

// Success 成功响应（带数据）
func Success(c *gin.Context, data any) {
	res := Response{
		Code: CodeSuccess,
		Msg:  "success",
		Data: data,
	}
	c.JSON(http.StatusOK, res)
}

// SuccessMsg 成功响应（无数据）
func SuccessMsg(c *gin.Context, msg string) {
	c.JSON(http.StatusOK, Response{
		Code: CodeSuccess,
		Msg:  msg,
	})
}

// Error 错误响应
func Error(c *gin.Context, msg string) {
	c.JSON(http.StatusOK, Response{
		Code: CodeError,
		Msg:  msg,
	})
}

// Fail 自定义错误码响应
func Fail(c *gin.Context, code int, msg string) {
	c.JSON(http.StatusOK, Response{
		Code: code,
		Msg:  msg,
	})
}

// FailWithCode 使用 ErrorCode 响应错误
func FailWithCode(c *gin.Context, ec errorcode.ErrorCode) {
	c.JSON(http.StatusOK, Response{
		Code: ec.Code,
		Msg:  ec.Msg,
	})
}

//go:embed template/*.html
var errorPageFS embed.FS

// HTML 返回自定义 HTML 错误页面（支持模板变量）
func HTML(c *gin.Context, status int, file string, data map[string]interface{}) {
	tmplBytes, err := errorPageFS.ReadFile("template/" + file)
	if err != nil {
		c.String(status, "读取页面失败: %v", err)
		return
	}

	tmpl, err := template.New("error").Parse(string(tmplBytes))
	if err != nil {
		c.String(status, "模板解析失败: %v", err)
		return
	}

	c.Status(status)
	c.Header("Content-Type", "text/html; charset=utf-8")
	if err := tmpl.Execute(c.Writer, data); err != nil {
		c.String(status, "模板渲染失败: %v", err)
	}
}

func NotFound(c *gin.Context, message string) {
	HTML(c, http.StatusNotFound, "404.html", map[string]interface{}{
		"Message": message,
	})
}
