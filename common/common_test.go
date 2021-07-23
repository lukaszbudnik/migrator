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

func TestFindNthIndex(t *testing.T) {
	indx := FindNthIndex("https://lukaszbudniktest.blob.core.windows.net/mycontainer/prod/artefacts", '/', 4)
	assert.Equal(t, 58, indx)
}

func TestFindNthIndexNotFound(t *testing.T) {
	indx := FindNthIndex("https://lukaszbudniktest.blob.core.windows.net/mycontainer", '/', 4)
	assert.Equal(t, -1, indx)
}
