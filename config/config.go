package config

import (
	"fmt"
	"gopkg.in/validator.v2"
	"gopkg.in/yaml.v2"
	"io/ioutil"
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

func FromFile(configFileName string) *Config {
	contents, err := ioutil.ReadFile(configFileName)

	if err != nil {
		panic(fmt.Sprintf("Could not read config file ==> %v", err))
	}

	return FromBytes(contents)
}

func FromBytes(contents []byte) *Config {
	var config Config

	if err := yaml.Unmarshal(contents, &config); err != nil {
		panic(fmt.Sprintf("Could not parse config file ==> %v", err))
	}

	if err := validator.Validate(config); err != nil {
		panic(fmt.Sprintf("Could not validate config file ==> %v", err))
	}

	return &config
}
