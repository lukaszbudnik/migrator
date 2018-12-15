package loader

import (
	"github.com/lukaszbudnik/migrator/config"
	"github.com/lukaszbudnik/migrator/types"
)

// Loader interface abstracts all loading operations performed by migrator
type Loader interface {
	GetDiskMigrations() ([]types.Migration, error)
}

// NewLoader returns new instance of Loader, currently DiskLoader is available
func NewLoader(config *config.Config) Loader {
	return &DiskLoader{config}
}
