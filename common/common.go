package common

import (
	"context"
	"fmt"
	"log"
	"path/filepath"
	"runtime"
)

const (
	panicLevel = "PANIC"
	errorLevel = "ERROR"
	warnLevel  = "WARN"
	infoLevel  = "INFO"
	debugLevel = "DEBUG"
)

// RequestIDKey is used together with context for setting/getting X-Request-ID
type RequestIDKey struct{}

// LogLevel
type LogLevelKey struct{}

// LogError logs error message
func LogError(ctx context.Context, format string, a ...interface{}) string {
	return logLevel(ctx, errorLevel, format, a...)
}

// LogWarn logs warning message
func LogWarn(ctx context.Context, format string, a ...interface{}) string {
	return logLevel(ctx, warnLevel, format, a...)
}

// LogInfo logs info message
func LogInfo(ctx context.Context, format string, a ...interface{}) string {
	return logLevel(ctx, infoLevel, format, a...)
}

// LogDebug logs debug message
func LogDebug(ctx context.Context, format string, a ...interface{}) string {
	return logLevel(ctx, debugLevel, format, a...)
}

// LogPanic logs error message
func LogPanic(ctx context.Context, format string, a ...interface{}) string {
	return logLevel(ctx, panicLevel, format, a...)
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

	logLevel := fmt.Sprintf("%v", ctx.Value(LogLevelKey{}))

	if shouldLogMessage(logLevel, level) {
		requestID := ctx.Value(RequestIDKey{})
		message := fmt.Sprintf(format, a...)
		_, file, line, _ := runtime.Caller(2)
		filename := filepath.Base(file)

		log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds | log.LUTC)
		log.Printf("[%v:%v] %v requestId=%v %v", filename, line, level, requestID, message)

		return message
	}

	return ""
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

func shouldLogMessage(configLogLevel, targetLevel string) bool {
	// if configLogLevel and targetLevel match then log
	if configLogLevel == targetLevel {
		return true
	}
	// if configLogLevel is debug then all messages are logged no need to check targetLevel
	if configLogLevel == debugLevel {
		return true
	}
	// if configLogLevel not set then INFO is assumed
	// if INFO then all levels should log except of debug
	if (len(configLogLevel) == 0 || configLogLevel == infoLevel) && targetLevel != debugLevel {
		return true
	}

	// if logLevel is ERROR then only ERROR and PANIC are logged
	// ERROR is covered in the beginning of method so need to check only Panic level
	if configLogLevel == errorLevel && targetLevel == panicLevel {
		return true
	}

	// if logLevel is WARN then WARN, ERROR and PANIC are logged
	if configLogLevel == warnLevel && (targetLevel == errorLevel || targetLevel == panicLevel) {
		return true
	}

	return false
}
