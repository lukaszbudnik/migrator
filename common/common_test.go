package common

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func newTestContext() context.Context {
	ctx := context.TODO()
	ctx = context.WithValue(ctx, RequestIDKey{}, "123")
	// log level empty = default log level = INFO
	ctx = context.WithValue(ctx, LogLevelKey{}, "")
	return ctx
}

func newTestContextWithDebugLogLevel() context.Context {
	ctx := newTestContext()
	ctx = context.WithValue(ctx, LogLevelKey{}, debugLevel)
	return ctx
}

func TestLogDebugSkip(t *testing.T) {
	// DEBUG message will be skipped, as the default log level is INFO
	message := LogDebug(newTestContext(), "success")
	assert.Empty(t, message)
}

func TestLogDebug(t *testing.T) {
	// DEBUG message will be returned, as the log level is set to DEBUG
	message := LogDebug(newTestContextWithDebugLogLevel(), "success")
	assert.Equal(t, "success", message)
}

func TestLogInfo(t *testing.T) {
	message := LogInfo(newTestContext(), "success")
	assert.Equal(t, "success", message)
}

func TestLogError(t *testing.T) {
	message := LogError(newTestContext(), "param=%v", 123)
	assert.Equal(t, "param=123", message)
}

func TestLogPanic(t *testing.T) {
	message := LogPanic(newTestContext(), "param=%v", 123456)
	assert.Equal(t, "param=123456", message)
}

func TestLog(t *testing.T) {
	message := Log("INFO", "param=%v", 456)
	assert.Equal(t, "param=456", message)
}

func TestFindNthIndex(t *testing.T) {
	indx := FindNthIndex("https://lukaszbudniktest.blob.core.windows.net/mycontainer/prod/artefacts", '/', 4)
	assert.Equal(t, 58, indx)
}

func TestFindNthIndexNotFound(t *testing.T) {
	indx := FindNthIndex("https://lukaszbudniktest.blob.core.windows.net/mycontainer", '/', 4)
	assert.Equal(t, -1, indx)
}

func TestShouldLogMessage(t *testing.T) {
	// default logLevel is info, should log all except of debug
	assert.False(t, shouldLogMessage("", debugLevel))
	assert.True(t, shouldLogMessage("", infoLevel))
	assert.True(t, shouldLogMessage("", errorLevel))
	assert.True(t, shouldLogMessage("", panicLevel))

	// debug logLevel logs all
	assert.True(t, shouldLogMessage(debugLevel, debugLevel))
	assert.True(t, shouldLogMessage(debugLevel, infoLevel))
	assert.True(t, shouldLogMessage(debugLevel, errorLevel))
	assert.True(t, shouldLogMessage(debugLevel, panicLevel))

	// info logLevel logs all except of debug
	assert.False(t, shouldLogMessage(infoLevel, debugLevel))
	assert.True(t, shouldLogMessage(infoLevel, infoLevel))
	assert.True(t, shouldLogMessage(infoLevel, errorLevel))
	assert.True(t, shouldLogMessage(infoLevel, panicLevel))

	// error logLevel logs only error or panic
	assert.False(t, shouldLogMessage(errorLevel, debugLevel))
	assert.False(t, shouldLogMessage(errorLevel, infoLevel))
	assert.True(t, shouldLogMessage(errorLevel, errorLevel))
	assert.True(t, shouldLogMessage(errorLevel, panicLevel))

	// panic logLevel logs only panic
	assert.False(t, shouldLogMessage(panicLevel, debugLevel))
	assert.False(t, shouldLogMessage(panicLevel, infoLevel))
	assert.False(t, shouldLogMessage(panicLevel, errorLevel))
	assert.True(t, shouldLogMessage(panicLevel, panicLevel))
}
