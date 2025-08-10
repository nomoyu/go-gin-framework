package db

import "sync"

var (
	modelsMu sync.RWMutex
	models   []any
)

// 仅注册：写锁/Unlock
func RegisterModel(ms ...any) {
	modelsMu.Lock()
	models = append(models, ms...)
	modelsMu.Unlock()
}

// 读取快照：读锁/RUnlock
func RegisteredModels() []any {
	modelsMu.RLock()
	defer modelsMu.RUnlock()
	out := make([]any, len(models))
	copy(out, models)
	return out
}

// 判空：读锁/RUnlock
func HasRegisteredModels() bool {
	modelsMu.RLock()
	defer modelsMu.RUnlock()
	return len(models) > 0
}
