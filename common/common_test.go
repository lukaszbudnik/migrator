package common

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func newTestContext() context.Context {
	ctx := context.TODO()
	ctx = context.WithValue(ctx, RequestIDKey{}, "123")
	return ctx
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
