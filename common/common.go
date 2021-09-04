package common

import (
	"context"
	"fmt"
	"log"
	"runtime"
)

// RequestIDKey is used together with context for setting/getting X-Request-ID
type RequestIDKey struct{}

// LogError logs error message
func LogError(ctx context.Context, format string, a ...interface{}) string {
	return logLevel(ctx, "ERROR", format, a...)
}

// LogInfo logs info message
func LogInfo(ctx context.Context, format string, a ...interface{}) string {
	return logLevel(ctx, "INFO", format, a...)
}

// LogPanic logs error message
func LogPanic(ctx context.Context, format string, a ...interface{}) string {
	return logLevel(ctx, "PANIC", format, a...)
}

// Log logs message with a given level with no request context
func Log(level string, format string, a ...interface{}) string {
	_, file, line, _ := runtime.Caller(2)

	message := fmt.Sprintf(format, a...)

	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds | log.LUTC)
	log.Printf("[%v:%v] %v %v", file, line, level, message)
	return message
}

func logLevel(ctx context.Context, level string, format string, a ...interface{}) string {
	_, file, line, _ := runtime.Caller(2)

	requestID := ctx.Value(RequestIDKey{})
	message := fmt.Sprintf(format, a...)

	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds | log.LUTC)
	log.Printf("[%v:%v] %v requestId=%v %v", file, line, level, requestID, message)
	return message
}

// FindNthIndex finds index of nth occurance of a character c in string str
func FindNthIndex(str string, c byte, n int) int {
	occur := 0
	for i := 0; i < len(str); i++ {
		if str[i] == c {
			occur++
		}
		if occur == n {
			return i
		}
	}
	return -1
}
