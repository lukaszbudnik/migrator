package common

import (
	"context"
	"fmt"
	"log"
)

// RequestIDKey is used together with context for setting/getting X-Request-Id
type RequestIDKey struct{}

// ActionKey is used together with context for setting/getting current action
type ActionKey struct{}

func logLevel(ctx context.Context, level string, format string, a ...interface{}) string {
	requestID := ctx.Value(RequestIDKey{})
	action := ctx.Value(ActionKey{})
	message := fmt.Sprintf(format, a...)
	log.Printf("%v %v [%v] - %v", level, action, requestID, message)
	return message
}

// LogError logs error message
func LogError(ctx context.Context, format string, a ...interface{}) string {
	return logLevel(ctx, "ERROR", format, a...)
}

// LogInfo logs info message
func LogInfo(ctx context.Context, format string, a ...interface{}) string {
	return logLevel(ctx, "INFO", format, a...)
}

// LogPanic logs error message and panics
func LogPanic(ctx context.Context, format string, a ...interface{}) string {
	message := logLevel(ctx, "PANIC", format, a...)
	panic(message)
}
