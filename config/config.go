package config

import (
	"io/ioutil"
	"log"
	"os"
	"strings"

	"gopkg.in/validator.v2"
	"gopkg.in/yaml.v2"
)

// Config represents Migrator's yaml configuration file
type Config struct {
	BaseDir           string   `yaml:"baseDir" validate:"nonzero"`
	Driver            string   `yaml:"driver" validate:"nonzero"`
	DataSource        string   `yaml:"dataSource" validate:"nonzero"`
	TenantSelectSQL   string   `yaml:"tenantSelectSQL,omitempty"`
	TenantInsertSQL   string   `yaml:"tenantInsertSQL,omitempty"`
	SchemaPlaceHolder string   `yaml:"schemaPlaceHolder,omitempty"`
	SingleSchemas     []string `yaml:"singleSchemas" validate:"min=1"`
	TenantSchemas     []string `yaml:"tenantSchemas,omitempty"`
	Port              string   `yaml:"port,omitempty"`
	WebHookURL        string   `yaml:"webHookURL,omitempty"`
	WebHookTemplate   string   `yaml:"webHookTemplate,omitempty"`
	WebHookHeaders    []string `yaml:"webHookHeaders,omitempty"`
}

func (config Config) String() string {
	c, _ := yaml.Marshal(config)
	return strings.TrimSpace(string(c))
}

// FromFile reads config from file which name is passed as an argument
func FromFile(configFileName string) (*Config, error) {
	contents, err := ioutil.ReadFile(configFileName)

	if err != nil {
		log.Printf("Could not read config file ==> %v", err)
		return nil, err
	}

	return FromBytes(contents)
}

// FromBytes reads config from raw bytes passed as an argument
func FromBytes(contents []byte) (*Config, error) {
	var config Config

	if err := yaml.Unmarshal(contents, &config); err != nil {
		log.Printf("Could not parse config file ==> %v", err)
		return nil, err
	}

	if err := validator.Validate(config); err != nil {
		log.Printf("Could not validate config file ==> %v", err)
		return nil, err
	}

	substituteEnvVariables(&config)

	return &config, nil
}

func substituteEnvVariables(config *Config) {
	if strings.HasPrefix(config.BaseDir, "$") {
		config.BaseDir = os.Getenv(config.BaseDir[1:])
	}
	if strings.HasPrefix(config.Driver, "$") {
		config.Driver = os.Getenv(config.Driver[1:])
	}
	if strings.HasPrefix(config.DataSource, "$") {
		config.DataSource = os.Getenv(config.DataSource[1:])
	}
	if strings.HasPrefix(config.TenantSelectSQL, "$") {
		config.TenantSelectSQL = os.Getenv(config.TenantSelectSQL[1:])
	}
	if strings.HasPrefix(config.TenantInsertSQL, "$") {
		config.TenantInsertSQL = os.Getenv(config.TenantInsertSQL[1:])
	}
	if strings.HasPrefix(config.Port, "$") {
		config.Port = os.Getenv(config.Port[1:])
	}
	if strings.HasPrefix(config.WebHookURL, "$") {
		config.WebHookURL = os.Getenv(config.WebHookURL[1:])
	}
	if strings.HasPrefix(config.WebHookTemplate, "$") {
		config.WebHookTemplate = os.Getenv(config.WebHookTemplate[1:])
	}
	if strings.HasPrefix(config.SchemaPlaceHolder, "$") {
		config.SchemaPlaceHolder = os.Getenv(config.SchemaPlaceHolder[1:])
	}

}
