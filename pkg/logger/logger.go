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
	"strings"
	"time"
)

var Log *zap.Logger
var sugar *zap.SugaredLogger

func InitLogger() {
	logPath := config.Conf.Log.Path
	level := config.Conf.Log.Level

	_ = os.MkdirAll(logPath, os.ModePerm)
	logFile := filepath.Join(logPath, time.Now().Format("2006-01-02")+".log")

	var zapLevel zapcore.Level
	if err := zapLevel.UnmarshalText([]byte(level)); err != nil {
		zapLevel = zapcore.InfoLevel
	}

	encoderConfig := zapcore.EncoderConfig{
		TimeKey:    "time",
		LevelKey:   "level",
		MessageKey: "msg",
		CallerKey:  "", // 不输出 caller
		EncodeTime: zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05.000"),
		EncodeLevel: func(l zapcore.Level, enc zapcore.PrimitiveArrayEncoder) {
			enc.AppendString(fmt.Sprintf("%-5s", strings.ToUpper(l.String())))
		},
	}

	consoleEncoder := zapcore.NewConsoleEncoder(encoderConfig)
	fileEncoder := zapcore.NewConsoleEncoder(encoderConfig)

	core := zapcore.NewTee(
		zapcore.NewCore(consoleEncoder, zapcore.Lock(os.Stdout), zapLevel),
		zapcore.NewCore(fileEncoder, zapcore.AddSync(&lumberjackWriter{filename: logFile}), zapLevel),
	)

	Log = zap.New(core)
	sugar = Log.Sugar()

	Log.Info("✅ 日志初始化完成")
}

// lumberjackWriter 模拟日志文件写入器（可替换为滚动日志）
type lumberjackWriter struct {
	filename string
}

func (l *lumberjackWriter) Write(p []byte) (n int, err error) {
	f, err := os.OpenFile(l.filename, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return 0, err
	}
	defer f.Close()
	return f.Write(p)
}

// InitLoggerWithConfig 可被 nomoyu.WithLog() 调用
func InitLoggerWithConfig(logPath string, level string) {
	if logPath == "" {
		logPath = "./logs"
	}
	if level == "" {
		level = "info"
	}

	_ = os.MkdirAll(logPath, os.ModePerm)
	logFile := filepath.Join(logPath, time.Now().Format("2006-01-02")+".log")

	var zapLevel zapcore.Level
	if err := zapLevel.UnmarshalText([]byte(level)); err != nil {
		zapLevel = zapcore.InfoLevel
	}

	encoderConfig := zapcore.EncoderConfig{
		TimeKey:    "time",
		LevelKey:   "level",
		MessageKey: "msg",
		CallerKey:  "",
		EncodeTime: zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05.000"),
		EncodeLevel: func(l zapcore.Level, enc zapcore.PrimitiveArrayEncoder) {
			enc.AppendString(fmt.Sprintf("%-5s", strings.ToUpper(l.String())))
		},
	}

	consoleEncoder := zapcore.NewConsoleEncoder(encoderConfig)
	fileEncoder := zapcore.NewConsoleEncoder(encoderConfig)

	core := zapcore.NewTee(
		zapcore.NewCore(consoleEncoder, zapcore.Lock(os.Stdout), zapLevel),
		zapcore.NewCore(fileEncoder, zapcore.AddSync(&lumberjackWriter{filename: logFile}), zapLevel),
	)

	Log = zap.New(core)
	sugar = Log.Sugar()
}

func ensureLogger() bool {
	if Log == nil {
		fmt.Fprintln(os.Stderr, "[LOGGER ERROR] logger 未初始化，请在框架启动时调用 nomoyu.WithLog(logPath, level)")
		return false
	}
	return true
}

// -----------------------------
// 简写方法（不带 context）
// -----------------------------

func Info(msg string, fields ...zap.Field) {
	Log.Info(msg, fields...)
}

func Warn(msg string, fields ...zap.Field) {
	Log.Warn(msg, fields...)
}

func Error(msg string, fields ...zap.Field) {
	Log.Error(msg, fields...)
}

func Infof(format string, args ...any) {
	sugar.Infof(format, args...)
}

func Warnf(format string, args ...any) {
	sugar.Warnf(format, args...)
}

func Errorf(format string, args ...any) {
	sugar.Errorf(format, args...)
}

// -----------------------------
// 带 traceID 的输出
// -----------------------------

func InfoWithTrace(ctx context.Context, msg string) {
	traceID := trace.GetTraceID(ctx)
	sugar.Infof("%-36s  %s", traceID, msg)
}

func ErrorWithTrace(ctx context.Context, msg string) {
	traceID := trace.GetTraceID(ctx)
	sugar.Errorf("%-36s  %s", traceID, msg)
}

func WarnWithTrace(ctx context.Context, msg string) {
	traceID := trace.GetTraceID(ctx)
	sugar.Warnf("%-36s  %s", traceID, msg)
}
