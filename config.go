package main

import (
	_ "fmt"
	"gopkg.in/validator.v2"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"strings"
)

// Config represents Migrator's yaml configuration file
type Config struct {
	BaseDir       string   `yaml:"baseDir" validate:"nonzero"`
	Driver        string   `yaml:"driver" validate:"nonzero"`
	DataSource    string   `yaml:"dataSource" validate:"nonzero"`
	TenantsSQL    string   `yaml:"tenantsSql"`
	SingleSchemas []string `yaml:"singleSchemas" validate:"min=1"`
	TenantSchemas []string `yaml:"tenantSchemas"`
}

func (config Config) String() string {
	c, _ := yaml.Marshal(config)
	return strings.TrimSpace(string(c))
}

func readConfigFromFile(configFile string) *Config {
	contents, err := ioutil.ReadFile(configFile)

	if err != nil {
		log.Panicf("Could not read config file ==> %v", err)
	}

	return readConfigFromBytes(contents)
}

func readConfigFromBytes(contents []byte) *Config {
	var config Config

	if err := yaml.Unmarshal(contents, &config); err != nil {
		log.Panicf("Could not parse config file ==> %v", err)
	}

	if err := validator.Validate(config); err != nil {
		log.Panicf("Could not validate config file ==> %v", err)
	}

	return &config
}
