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
	message := LogInfo(newTestContext(), "result=success")
	assert.Equal(t, "result=success", message)
}

func TestLogError(t *testing.T) {
	message := LogError(newTestContext(), "param=%v", 123)
	assert.Equal(t, "param=123", message)
}

func TestLogPanic(t *testing.T) {
	assert.Panics(t, func() {
		LogPanic(newTestContext(), "param=%v", 123)
	})
}
