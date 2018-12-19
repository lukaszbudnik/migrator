package common

import (
	"context"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func newTestContext() context.Context {
	pc, _, _, _ := runtime.Caller(1)
	details := runtime.FuncForPC(pc)

	ctx := context.TODO()
	ctx = context.WithValue(ctx, RequestIDKey{}, "123")
	ctx = context.WithValue(ctx, ActionKey{}, strings.Replace(details.Name(), "github.com/lukaszbudnik/migrator/common.", "", -1))
	return ctx
}

func TestLogInfo(t *testing.T) {
	message := LogInfo(newTestContext(), "format no params")
	assert.Equal(t, "format no params", message)
}

func TestLogError(t *testing.T) {
	message := LogError(newTestContext(), "format no params: %v", 123)
	assert.Equal(t, "format no params: 123", message)
}

func TestLogPanic(t *testing.T) {
	assert.Panics(t, func() {
		LogPanic(newTestContext(), "format no params: %v", 123)
	})
}
