package loader

import (
	"context"
	"testing"

	"github.com/lukaszbudnik/migrator/config"
	"github.com/stretchr/testify/assert"
)

func TestNewDiskLoader(t *testing.T) {
	config := &config.Config{
		BaseDir: "/path/to/baseDir",
	}
	loader := New(context.TODO(), config)
	assert.IsType(t, &diskLoader{}, loader)
}

func TestNewAzureBlobLoader(t *testing.T) {
	config := &config.Config{
		BaseDir: "https://lukaszbudniktest.blob.core.windows.net/mycontainer",
	}
	loader := New(context.TODO(), config)
	assert.IsType(t, &azureBlobLoader{}, loader)
}

func TestNewS3Loader(t *testing.T) {
	config := &config.Config{
		BaseDir: "s3://lukaszbudniktest-bucket",
	}
	loader := New(context.TODO(), config)
	assert.IsType(t, &s3Loader{}, loader)
}
