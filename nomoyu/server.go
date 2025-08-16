package nomoyu

import (
	"context"
	"fmt"
	"github.com/nomoyu/go-gin-framework/internal/middleware"
	"github.com/nomoyu/go-gin-framework/pkg/config"
	"github.com/nomoyu/go-gin-framework/pkg/logger"
	"net/http"
	"os"
	"os/signal"
	"sync/atomic"
	"syscall"
	"time"
)

type ServerOption struct {
	Port     string
	FromUser bool
}

func (a *App) Run(addr ...string) {
	// 注册模块
	for _, m := range a.modules {
		m.Register(a.engine)
	}

	// 注册路由分组
	for _, group := range a.routes {
		g := a.engine.Group(group.prefix)
		// ✅ 如果启用了权限认证模块并且该路由声明了 RequireAuth
		if group.requireAuth && a.authOption != nil {
			g.Use(middleware.AuthMiddleware(a.authOption.Strategy))
		}
		if len(group.middleware) > 0 {
			g.Use(group.middleware...)
		}
		for _, register := range group.routes {
			register(g)
		}
	}

	// 端口优先级：传参 > 配置 > 默认
	port := ":3303"
	if len(addr) > 0 && addr[0] != "" {
		port = addr[0]
	} else if config.Conf.Server.Port != 0 {
		port = fmt.Sprintf(":%d", config.Conf.Server.Port)
	}

	// ✅ 默认停机超时（可再做 WithShutdownTimeout 扩展）
	if a.shutdownTimeout == 0 {
		a.shutdownTimeout = 10 * time.Second
	}

	// ✅ 关机防护中间件（建议放最前或日志/trace之后）
	a.engine.Use(a.shutdownGuard())

	// 用 http.Server 承载，支持优雅停机
	a.httpServer = &http.Server{
		Addr:         port,
		Handler:      a.engine,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// 启动服务（异步）
	errCh := make(chan error, 1)
	go func() {
		logger.Infof("server listening on %s", port)
		if err := a.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- err
			return
		}
		errCh <- nil
	}()

	// 捕获信号（Ctrl+C / 容器 SIGTERM）
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	select {
	case sig := <-quit:
		logger.Warnf("received signal: %v, start graceful shutdown...", sig)
	case err := <-errCh:
		if err != nil {
			logger.Errorf("http server error: %v", err)
		} else {
			logger.Infof("http server exited")
		}
	}

	// 标记进入关机，拒绝新请求
	atomic.StoreInt32(&a.shutting, 1)

	// 关机超时上下文
	ctx, cancel := context.WithTimeout(context.Background(), a.shutdownTimeout)
	defer cancel()

	// 优雅关闭 HTTP
	if a.httpServer != nil {
		if err := a.httpServer.Shutdown(ctx); err != nil {
			logger.Errorf("http server shutdown error: %v", err)
		} else {
			logger.Infof("http server shutdown gracefully")
		}
	}

	// 依次执行清理钩子（倒序更合理）
	for i := len(a.shutdownHooks) - 1; i >= 0; i-- {
		if err := a.shutdownHooks[i](ctx); err != nil {
			logger.Warnf("shutdown hook error: %v", err)
		}
	}

	logger.Infof("graceful shutdown done, bye 👋")
}
