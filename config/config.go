package config

import (
	"os"
	"reflect"
	"strings"

	"gopkg.in/go-playground/validator.v9"
	"gopkg.in/yaml.v3"
)

// Config represents Migrator's yaml configuration file
type Config struct {
	BaseLocation      string   `yaml:"baseLocation" validate:"required"`
	Driver            string   `yaml:"driver" validate:"required"`
	DataSource        string   `yaml:"dataSource" validate:"required"`
	TenantSelect      string   `yaml:"tenantSelect,omitempty"`
	TenantInsert      string   `yaml:"tenantInsert,omitempty"`
	TenantSelectSQL   string   `yaml:"tenantSelectSQL,omitempty"` // Deprecated: use TenantSelect instead
	TenantInsertSQL   string   `yaml:"tenantInsertSQL,omitempty"` // Deprecated: use TenantInsert instead
	SchemaPlaceHolder string   `yaml:"schemaPlaceHolder,omitempty"`
	SingleMigrations  []string `yaml:"singleMigrations" validate:"min=1"`
	TenantMigrations  []string `yaml:"tenantMigrations,omitempty"`
	SingleScripts     []string `yaml:"singleScripts,omitempty"`
	TenantScripts     []string `yaml:"tenantScripts,omitempty"`
	Port              string   `yaml:"port,omitempty"`
	PathPrefix        string   `yaml:"pathPrefix,omitempty"`
	WebHookURL        string   `yaml:"webHookURL,omitempty"`
	WebHookHeaders    []string `yaml:"webHookHeaders,omitempty"`
	WebHookTemplate   string   `yaml:"webHookTemplate,omitempty"`
	LogLevel          string   `yaml:"logLevel,omitempty" validate:"logLevel"`
}

// GetTenantSelect returns tenant select query/statement with backward compatibility
func (c *Config) GetTenantSelect() string {
	// New field takes precedence
	if c.TenantSelect != "" {
		return c.TenantSelect
	}
	// Fall back to deprecated field with warning
	if c.TenantSelectSQL != "" {
		// Note: We can't use common.LogWarn here as it requires context
		// The warning will be logged when the config is actually used
		return c.TenantSelectSQL
	}
	return ""
}

// GetTenantInsert returns tenant insert query/statement with backward compatibility
func (c *Config) GetTenantInsert() string {
	// New field takes precedence
	if c.TenantInsert != "" {
		return c.TenantInsert
	}
	// Fall back to deprecated field with warning
	if c.TenantInsertSQL != "" {
		// Note: We can't use common.LogWarn here as it requires context
		// The warning will be logged when the config is actually used
		return c.TenantInsertSQL
	}
	return ""
}

// IsUsingDeprecatedTenantSelectSQL returns true if deprecated field is being used
func (c *Config) IsUsingDeprecatedTenantSelectSQL() bool {
	return c.TenantSelect == "" && c.TenantSelectSQL != ""
}

// IsUsingDeprecatedTenantInsertSQL returns true if deprecated field is being used
func (c *Config) IsUsingDeprecatedTenantInsertSQL() bool {
	return c.TenantInsert == "" && c.TenantInsertSQL != ""
}

func (config Config) String() string {
	c, _ := yaml.Marshal(config)
	return strings.TrimSpace(string(c))
}

// FromFile reads config from file which name is passed as an argument
func FromFile(configFileName string) (*Config, error) {
	contents, err := os.ReadFile(configFileName)

	if err != nil {
		return nil, err
	}

	return FromBytes(contents)
}

// FromBytes reads config from raw bytes passed as an argument
func FromBytes(contents []byte) (*Config, error) {
	var config Config

	if err := yaml.Unmarshal(contents, &config); err != nil {
		return nil, err
	}

	validate := validator.New()
	validate.RegisterValidation("logLevel", validateLogLevel)
	if err := validate.Struct(config); err != nil {
		return nil, err
	}

	substituteEnvVariables(&config)

	return &config, nil
}

func substituteEnvVariables(config *Config) {
	val := reflect.ValueOf(config).Elem()
	for i := 0; i < val.NumField(); i++ {
		valueField := val.Field(i)
		typeField := val.Type().Field(i)
		if val.CanAddr() && val.CanSet() {
			switch typeField.Type.Kind() {
			case reflect.String:
				s := valueField.Interface().(string)
				s = substituteEnvVariable(s)
				valueField.SetString(s)
			case reflect.Slice:
				ss := valueField.Interface().([]string)
				for i := range ss {
					ss[i] = substituteEnvVariable(ss[i])
				}
				valueField.Set(reflect.ValueOf(ss))
			}
		}
	}
}

func substituteEnvVariable(s string) string {
	start := strings.Index(s, "${")
	end := strings.Index(s, "}")
	if start > -1 && end > start && len(s) > end {
		after := s[0:start] + os.Getenv(s[start+2:end]) + s[end+1:]
		return substituteEnvVariable(after)
	}
	return s
}

func validateLogLevel(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	return value == "" || value == "DEBUG" || value == "INFO" || value == "ERROR" || value == "PANIC"
}
