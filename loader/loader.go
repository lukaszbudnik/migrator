package loader

import (
	"context"

	"github.com/lukaszbudnik/migrator/config"
	"github.com/lukaszbudnik/migrator/types"
)

// Loader interface abstracts all loading operations performed by migrator
type Loader interface {
	GetSourceMigrations() []types.Migration
}

// LoaderFactory is a factory method for creating Loader instance
type Factory func(context.Context, *config.Config) Loader

// New returns new instance of Loader, currently DiskLoader is available
func New(ctx context.Context, config *config.Config) Loader {
	return &diskLoader{ctx, config}
}
