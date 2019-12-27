package common

import (
	"context"
	"fmt"
	"log"
	"runtime"
)

// RequestIDKey is used together with context for setting/getting X-Request-Id
type RequestIDKey struct{}

// ActionKey is used together with context for setting/getting current action
type ActionKey struct{}

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

func logLevel(ctx context.Context, level string, format string, a ...interface{}) string {
	_, file, line, _ := runtime.Caller(2)

	requestID := ctx.Value(RequestIDKey{})
	message := fmt.Sprintf(format, a...)

	log.SetFlags(log.LstdFlags | log.LUTC | log.Lmicroseconds)
	log.Printf("[%v:%v] %v requestId=%v %v", file, line, level, requestID, message)
	return message
}
