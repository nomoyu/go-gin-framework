| 模块           | 说明                                            |
| ------------ | --------------------------------------------- |
| `config`     | 使用 `viper` 读取配置文件，支持热更新和多环境配置切换               |
| `logger`     | 使用 `zap`，支持日志等级、格式、traceID，按日期/大小滚动日志         |
| `middleware` | JWT 校验、CORS、请求日志、限流器（如 token bucket）          |
| `response`   | 统一封装返回结构体（code、msg、data）                      |
| `model`      | 定义 GORM 结构体 + 关联关系                            |
| `dao`        | 数据访问封装（事务、分页、基础CRUD）                          |
| `service`    | 具体业务逻辑，实现 handler 与 dao 分离                    |
| `handler`    | HTTP 控制器，参数校验（推荐使用 `go-playground/validator`） |
| `router`     | Gin 的路由分组注册，例如 `/api/v1/users`                |
| `pkg`        | 常用工具：雪花ID生成器、密码加密、通用响应、时间格式化等                 |


# 🌐 nomoyu 路由模块说明文档（RouteGroup）

为了简化业务项目中 Gin 路由的注册方式，并保持结构清晰、支持分组、链式调用和中间件挂载，框架提供了一个 `nomoyu.RouteGroup` 的简洁封装。

---

## ✅ 基本用法

```go
import (
    "github.com/gin-gonic/gin"
    "github.com/nomoyu/go-gin-framework/nomoyu"
)

func Routes() nomoyu.RouteGroup {
    return nomoyu.NewGroup("/user").
        GET("/hello", helloHandler).
        POST("/login", loginHandler)
}
```

上面代码会注册以下两个接口：

| 方法  | 路由路径     | 处理函数      |
|-------|--------------|----------------|
| GET   | `/user/hello` | `helloHandler` |
| POST  | `/user/login` | `loginHandler` |

---

## 🧩 接口文档示例（Swagger）

框架默认支持 swagger 文档，只需在每个 handler 上添加注释：

```go
// helloHandler godoc
// @Summary     用户打招呼
// @Tags        用户模块
// @Produce     json
// @Success     200  {string}  string  "Hello!"
// @Router      /user/hello [get]
```

---

## 🛠 API 说明

### `NewGroup(prefix string)`

用于定义一个路由组前缀，例如 `/user`。

### 链式方法

| 方法           | 说明                      |
|----------------|---------------------------|
| `.GET(path, handler)`     | 注册 GET 路由       |
| `.POST(path, handler)`    | 注册 POST 路由      |
| `.PUT(path, handler)`     | 注册 PUT 路由       |
| `.DELETE(path, handler)`  | 注册 DELETE 路由    |
| `.Use(middleware...)`     | 挂载中间件到当前组  |

---

## 🔄 完整示例：ping 接口注册

```go
package api

import (
    "net/http"
    "github.com/gin-gonic/gin"
    "github.com/nomoyu/go-gin-framework/nomoyu"
)

// pingHandler godoc
// @Summary     健康检查接口
// @Description 测试服务存活状态
// @Tags        系统
// @Produce     json
// @Success     200  {string}  string  "pong"
// @Router      /ping [get]
func pingHandler(c *gin.Context) {
    c.String(http.StatusOK, "pong")
}

// PingRoutes 返回路由组
func PingRoutes() nomoyu.RouteGroup {
    return nomoyu.NewGroup("/").
        GET("/ping", pingHandler)
}
```

---

## 🎯 和 Gin 原生写法对比

| 原始 Gin 写法                                      | RouteGroup 封装写法                        |
|----------------------------------------------------|---------------------------------------------|
| `r.GET("/ping", handler)`                          | `NewGroup("/").GET("/ping", handler)`       |
| `group := r.Group("/user")`                        | `NewGroup("/user")`                         |
| `group.Use(AuthMiddleware())`                      | `.Use(AuthMiddleware())`                    |

---

## 📦 和 `nomoyu.Start()` 联动使用

在主程序中统一挂载所有路由组：

```go
func main() {
    app := nomoyu.Start().
        WithRoute(api.Routes(), api.PingRoutes())

    app.Run()
}
```

---

## 🔐 支持中间件注入

```go
return nomoyu.NewGroup("/admin").
    Use(AuthMiddleware()).
    GET("/dashboard", dashboardHandler)
```


# 📘 使用 Nomoyu 框架启用 Swagger 接口文档

本指南将帮助你在基于 Nomoyu 框架的 Go 项目中快速启用 Swagger API 文档。

---

## 🧩 一、安装 Swag 工具（只需一次）

```bash
go install github.com/swaggo/swag/cmd/swag@latest
```

确保 `$GOPATH/bin` 在环境变量中：

```bash
echo 'export PATH=$PATH:$(go env GOPATH)/bin' >> ~/.zshrc
source ~/.zshrc
```

---

## ⚙️ 二、配置文件添加 Swagger 配置

在 `configs/config.dev.yaml` 中添加：

```yaml
swagger:
  enabled: true
  route: /swagger/*any
```

---

## ✍️ 三、在 `main.go` 添加注释头 & 导入

```go
// @title           Nomoyu 示例接口文档
// @version         1.0
// @description     这是 nomoyu 自动生成的 Swagger 文档
// @host            localhost:9000
// @BasePath        /

package main

import (
    _ "github.com/yourname/test-nomoyu/docs"
    "github.com/nomoyu/go-gin-framework/nomoyu"
    "github.com/nomoyu/test-nomoyu/api"
)
```

---

## 🧱 四、启用 Swagger 模块

```go
nomoyu.Start().
    WithRoute(api.PingRoutes()).
    WithSwagger().
    Run()
```

---

## 📌 五、为接口添加注解（以 `/ping` 为例）

```go
// pingHandler godoc
// @Summary      健康检查接口
// @Description  用于测试服务是否存活
// @Tags         系统
// @Produce      json
// @Success      200 {string} string "pong"
// @Router       /ping [get]
func pingHandler(c *gin.Context) {
    c.String(200, "pong")
}
```

---

## 🏗️ 六、生成 Swagger 文档

运行：

```bash
swag init -g main.go
```

生成以下文件夹：

```
docs/
├── docs.go
├── swagger.json
└── swagger.yaml
```

---

## 🌐 七、访问 Swagger UI

访问：

```
http://localhost:9000/swagger/index.html
```

> 实际地址由配置文件中的 `swagger.route` 决定。

---

## ✅ 自动加载机制说明

| 场景                         | 是否启用 Swagger |
|------------------------------|------------------|
| 显式调用 `.WithSwagger()`    | ✅ 启用（优先级高） |
| 配置文件开启 `enabled: true` | ✅ 启用 |
| 都未配置                     | ❌ 不启用 |

---

## 🧯 常见问题排查

| 问题                                      | 原因和解决方法 |
|-------------------------------------------|----------------|
| `doc.json` 报错或找不到                   | 未导入 `docs` 包；未运行 `swag init` |
| Swagger 没显示                            | 配置文件未开启 `swagger.enabled` 或未调用 `.WithSwagger()` |
| Swagger 地址 404 或端口不对               | 检查端口 `server.port` 是否一致；检查配置 `route` |
| `swag` 命令不存在                         | 没有安装 swag 工具，请参考上文第一步 |

---

# 🔐 Auth 认证模块使用文档

本框架支持基于策略模式实现的可插拔认证模块，目前默认内置 `JWT` 认证支持，支持配置文件启用和链式调用注入，遵循约定优于配置设计，适配多种业务场景。

---

## 🧩 一、配置启用（推荐）

在配置文件中添加如下内容即可启用认证模块，无需手动调用：

```yaml
auth:
  enabled: true
  mode: "jwt"
  jwt:
    secret: "mySecretKey"
```

- `auth.enabled`: 是否启用认证模块
- `auth.mode`: 当前支持 `"jwt"`
- `jwt.secret`: 用于签名和解析 JWT 的密钥

---

## 🧬 二、链式调用启用（优先级高）

框架支持手动注入认证策略，优先级高于配置文件：

```go
import "github.com/nomoyu/go-gin-framework/internal/auth"

nomoyu.Start().
    WithAuth(&auth.JWTStrategy{Secret: "mySecretKey"}).
    WithRoute(...)
```

---

## 🛠 三、生成 Token（登录接口使用）

通过工具函数生成用户 Token，支持自定义 Claims 字段：

```go
import "github.com/nomoyu/go-gin-framework/internal/auth"

claims := map[string]interface{}{
    "id":    "123",
    "name":  "Alice",
    "email": "alice@example.com",
    "roles": []string{"admin", "user"},
}

token, err := auth.GenerateJWTWithMap(claims, "mySecretKey", 2*time.Hour)
```

---

## 🔍 四、请求时携带 Token

RequireAuth启动鉴权，没有默认不鉴权

```go
// Routes 注册用户模块相关路由
func Routes() nomoyu.RouteGroup {
	return nomoyu.NewGroup("/user").RequireAuth().
		GET("/hello", helloHandler)
}

// helloHandler godoc
// @Summary      用户打招呼接口
// @Description  用户访问问候接口，返回欢迎语
// @Tags         用户模块
// @Produce      json
// @Success      200  {object}  map[string]interface{}  "响应成功"
// @Router       /user/hello [get]
func helloHandler(c *gin.Context) {
	authInfo := auth.GetAuthInfo(c)
	if authInfo != nil {
		email, _ := authInfo["email"].(string)
		logger.Infof("当前用户 email：%s", email)
	}

	response.Success(c, "Hello User!")
}
```

调用受保护接口时在请求头添加：

```http
Authorization: Bearer <your-token>
```

---

## 🔑 五、中间件鉴权流程

框架自动注入 `AuthMiddleware`，实现如下：

```go
func AuthMiddleware(strategy AuthStrategy) gin.HandlerFunc {
    return func(c *gin.Context) {
        authInfo, err := strategy.Authenticate(c)
        if err != nil {
            response.Unauthorized(c, "权限校验失败")
            return
        }

        // 将认证信息写入上下文
        c.Set("AuthInfo", authInfo)
        c.Next()
    }
}
```

---

## 🧾 六、获取认证信息（通用封装）

框架封装了通用函数 `authutil.GetAuthInfo(c)`，从上下文中提取认证信息：

```go
import "github.com/nomoyu/go-gin-framework/pkg/authutil"

func HelloHandler(c *gin.Context) {
    authInfo := authutil.GetAuthInfo(c)
    if authInfo != nil {
        email := authInfo["email"]
        logger.Infof("当前用户 Email：%v", email)
    }

    response.Success(c, "Hello!")
}
```

---

## 🚫 七、访问未携带 Token 接口时

将自动拦截，并返回：

```json
{
  "code": 401,
  "message": "权限校验失败"
}
```

---

## 📦 八、未来扩展支持

支持按需扩展其他认证方式：

- OAuth2
- Session / Cookie
- 第三方认证平台（如企业微信、GitHub、钉钉等）

实现 `AuthStrategy` 接口即可：

```go
type AuthStrategy interface {
    Authenticate(c *gin.Context) (map[string]interface{}, error)
}
```
