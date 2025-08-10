package logger

import (
	"context"
	"fmt"
	"github.com/nomoyu/go-gin-framework/pkg/config"
	"github.com/nomoyu/go-gin-framework/pkg/trace"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	logApp   *zap.Logger
	sugar    *zap.SugaredLogger
	AtomicLv = zap.NewAtomicLevel()

	reqBind sync.Map // map[uint64]*zap.Logger（按 goroutine 绑定 traceID）
)

func InitLogger() { InitLoggerWithConfig(config.Conf.Log.Path, config.Conf.Log.Level) }

// Internal: used by framework middleware only.
func InitLoggerWithConfig(logPath, level string) {
	if logPath == "" {
		logPath = "./logs"
	}
	if level == "" {
		level = "info"
	}
	_ = os.MkdirAll(logPath, 0755)

	appFile := filepath.Join(logPath, time.Now().Format("2006-01-02")+".log")
	if err := AtomicLv.UnmarshalText([]byte(level)); err != nil {
		AtomicLv.SetLevel(zap.InfoLevel)
	}

	// --- 两套 Encoder：控制台彩色，文件纯净 ---
	consoleEncCfg := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "trace", // 第二列显示 traceId（logger name）
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stack",
		EncodeTime:     zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05.000"),
		EncodeLevel:    zapcore.CapitalColorLevelEncoder, // 👈 级别自带颜色
		EncodeCaller:   encodeCallerColor,                // 👈 caller 上色
		EncodeDuration: zapcore.StringDurationEncoder,
	}
	fileEncCfg := consoleEncCfg
	fileEncCfg.EncodeLevel = func(l zapcore.Level, enc zapcore.PrimitiveArrayEncoder) {
		enc.AppendString(fmt.Sprintf("%-5s", strings.ToUpper(l.String())))
	}
	fileEncCfg.EncodeCaller = encodeCallerPlain
	fileEncCfg.EncodeDuration = zapcore.StringDurationEncoder

	consoleEnc := zapcore.NewConsoleEncoder(consoleEncCfg)
	fileEnc := zapcore.NewConsoleEncoder(fileEncCfg)

	core := zapcore.NewTee(
		zapcore.NewCore(consoleEnc, zapcore.AddSync(os.Stdout), AtomicLv),      // 彩色控制台
		zapcore.NewCore(fileEnc, zapcore.AddSync(mustFile(appFile)), AtomicLv), // 纯净文件
	)

	logApp = zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1), zap.AddStacktrace(zap.DPanicLevel))
	sugar = logApp.Sugar()
	logApp.Info("✅ 日志初始化完成")
}

func mustFile(p string) zapcore.WriteSyncer {
	f, err := os.OpenFile(p, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}
	return zapcore.AddSync(f)
}

// —— caller 仅保留 “最后两级/文件.go:行号”，控制台上色，文件不加色 ——
// e.g. ping/ping_handler.go:14
func encodeCallerPlain(c zapcore.EntryCaller, enc zapcore.PrimitiveArrayEncoder) {
	// 当我们用 WithRouteColumn(route) 时，Line 被设为 0，直接输出 File（就是 route）
	if c.Line == 0 {
		enc.AppendString(c.File)
		return
	}
	enc.AppendString(short2(c.TrimmedPath()))
}
func encodeCallerColor(c zapcore.EntryCaller, enc zapcore.PrimitiveArrayEncoder) {
	const cyan = "\x1b[36m"
	const reset = "\x1b[0m"
	enc.AppendString(cyan + short2(c.TrimmedPath()) + reset)
}
func short2(path string) string {
	parts := strings.Split(path, "/")
	if len(parts) >= 2 {
		return parts[len(parts)-2] + "/" + parts[len(parts)-1]
	}
	return path
}

// BindTraceForRequest —— 绑定 traceId：把 traceId 放到 Name 列 ——
// 中间件里调用 bind/unbind，让后续 logger.Info 自动带 traceId
func BindTraceForRequest(traceID string) func() {
	l := logApp.Named(traceID)
	gid := curGID()
	reqBind.Store(gid, l)
	return func() { reqBind.Delete(gid) }
}

func CurLogger() *zap.Logger {
	if v, ok := reqBind.Load(curGID()); ok {
		if l, ok2 := v.(*zap.Logger); ok2 {
			return l
		}
	}
	if logApp != nil {
		return logApp
	}
	return zap.NewNop()
}
func CurSugar() *zap.SugaredLogger { return CurLogger().Sugar() }

func curGID() uint64 {
	var b [64]byte
	n := runtime.Stack(b[:], false)
	s := strings.TrimPrefix(string(b[:n]), "goroutine ")
	i := strings.IndexByte(s, ' ')
	if i < 1 {
		return 0
	}
	id, _ := strconv.ParseUint(s[:i], 10, 64)
	return id
}

// L —— 简写 API（都会用当前请求 logger） ——
func L() *zap.Logger        { return CurLogger() }
func S() *zap.SugaredLogger { return CurSugar() }

// WithTrace 兼容旧用法：手动加 trace
func WithTrace(ctx context.Context) *zap.SugaredLogger {
	tid := trace.GetTraceID(ctx)
	return logApp.Named(tid).Sugar()
}

// WithRouteColumn 返回一个仅“本次日志”生效的 logger，
// 会把 caller 列显示为指定的 route（如 "/ping"）
func WithRouteColumn(route string) *zap.Logger {
	return L().WithOptions(zap.WrapCore(func(core zapcore.Core) zapcore.Core {
		return &routeCallerCore{Core: core, route: route}
	}))
}

type routeCallerCore struct {
	zapcore.Core
	route string
}

func (c *routeCallerCore) Write(ent zapcore.Entry, fields []zapcore.Field) error {
	// 只在最终 Write 前覆盖 caller，确保赢过 AddCaller
	ent.Caller = zapcore.EntryCaller{
		Defined: true,
		File:    c.route, // 直接用 "/ping"
		Line:    0,
	}
	return c.Core.Write(ent, fields)
}

// With 其它方法保持默认转发
func (c *routeCallerCore) With(fields []zapcore.Field) zapcore.Core {
	return &routeCallerCore{Core: c.Core.With(fields), route: c.route}
}
func (c *routeCallerCore) Check(ent zapcore.Entry, ce *zapcore.CheckedEntry) *zapcore.CheckedEntry {
	if c.Enabled(ent.Level) {
		return ce.AddCore(ent, c)
	}
	return ce
}
