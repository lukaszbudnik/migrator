package loader

import (
	"context"
	"regexp"
	"sort"
	"strings"

	"github.com/lukaszbudnik/migrator/config"
	"github.com/lukaszbudnik/migrator/types"
)

// Loader interface abstracts all loading operations performed by migrator
type Loader interface {
	GetSourceMigrations() []types.Migration
	HealthCheck() error
}

// Factory is a factory method for creating Loader instance
type Factory func(context.Context, *config.Config) Loader

// New returns new instance of Loader
func New(ctx context.Context, config *config.Config) Loader {
	if strings.HasPrefix(config.BaseLocation, "s3://") {
		return &s3Loader{
			baseLoader:       baseLoader{ctx, config},
			clientFactory:    &defaultS3ClientFactory{},
			paginatorFactory: &defaultS3PaginatorFactory{},
		}
	}
	if matched, _ := regexp.Match(`^https://.*\.blob\.core\.windows\.net/.*`, []byte(config.BaseLocation)); matched {
		return &azureBlobLoader{baseLoader{ctx, config}}
	}
	return &diskLoader{baseLoader{ctx, config}}
}

// baseLoader is the base struct for implementing Loader interface
type baseLoader struct {
	ctx    context.Context
	config *config.Config
}

func (bl *baseLoader) sortMigrations(migrationsMap map[string][]types.Migration, migrations *[]types.Migration) {
	keys := make([]string, 0, len(migrationsMap))
	for key := range migrationsMap {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	for _, key := range keys {
		ms := migrationsMap[key]
		*migrations = append(*migrations, ms...)
	}
}
