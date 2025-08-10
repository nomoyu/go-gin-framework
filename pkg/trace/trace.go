package trace

import (
	"context"

	"github.com/google/uuid"
)

const TraceIDKey = "traceID"

func NewTraceID() string {
	return uuid.New().String()
}

// WithTraceID 设置 traceID 到上下文
func WithTraceID(ctx context.Context, traceID string) context.Context {
	return context.WithValue(ctx, TraceIDKey, traceID)
}

// GetTraceID 从上下文中获取 traceID
func GetTraceID(ctx context.Context) string {
	if traceID, ok := ctx.Value(TraceIDKey).(string); ok {
		return traceID
	}
	return ""
}
