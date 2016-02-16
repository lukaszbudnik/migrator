package main

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestReadConfigFromFile(t *testing.T) {

	config, err := readConfigFromFile("test/migrator.yaml")

	assert.Nil(t, err)
	assert.Equal(t, "test/migrations", config.SourceDir)
	assert.Equal(t, "test/migrations", config.SourceDir)
	assert.Equal(t, "postgres", config.Driver)
	assert.Equal(t, "user=postgres dbname=postgres host=192.168.99.100 port=55432 sslmode=disable", config.DataSource)
	assert.Equal(t, []string{"tenants"}, config.TenantSchemas)
	assert.Equal(t, []string{"public", "ref", "config"}, config.SingleSchemas)

}

func TestReadConfigFromEmptyFile(t *testing.T) {

	config, err := readConfigFromFile("test/empty.yaml")

	assert.Nil(t, config)
	assert.Error(t, err)

}

func TestReadConfigFromNonExistingFile(t *testing.T) {

	config, err := readConfigFromFile("abcxyz.yaml")

	assert.Nil(t, config)
	assert.Error(t, err)

}
