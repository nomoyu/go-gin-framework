package logger

import (
	log "github.com/nomoyu/go-gin-framework/internal/logger"
	"go.uber.org/zap"
)

func SetLevel(lvl string) error { return log.AtomicLv.UnmarshalText([]byte(lvl)) }
func GetLevel() string          { return log.AtomicLv.Level().String() }

func Debug(msg string, fields ...zap.Field) { log.CurLogger().Debug(msg, fields...) }
func Info(msg string, fields ...zap.Field)  { log.CurLogger().Info(msg, fields...) }
func Warn(msg string, fields ...zap.Field)  { log.CurLogger().Warn(msg, fields...) }
func Error(msg string, fields ...zap.Field) { log.CurLogger().Error(msg, fields...) }

func Debugf(format string, args ...any) { log.CurSugar().Debugf(format, args...) }
func Infof(format string, args ...any)  { log.CurSugar().Infof(format, args...) }
func Warnf(format string, args ...any)  { log.CurSugar().Warnf(format, args...) }
func Errorf(format string, args ...any) { log.CurSugar().Errorf(format, args...) }
