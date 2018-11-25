package loader

import (
	"github.com/lukaszbudnik/migrator/config"
	"github.com/lukaszbudnik/migrator/types"
	"github.com/lukaszbudnik/migrator/utils"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sort"
)

// DiskLoader is struct used for implementing Loader interface for loading migrations from disk
type DiskLoader struct {
	Config *config.Config
}

// GetMigrations loads all migrations from disk
func (dl *DiskLoader) GetMigrations() []types.Migration {
	dirs, err := ioutil.ReadDir(dl.Config.BaseDir)
	if err != nil {
		log.Panicf("Could not read migration base dir: %v", err)
	}

	singleSchemasDirs := dl.filterSchemaDirs(dirs, dl.Config.SingleSchemas)
	tenantSchemasDirs := dl.filterSchemaDirs(dirs, dl.Config.TenantSchemas)

	migrationsMap := make(map[string][]types.Migration)

	dl.readMigrationsFromSchemaDirs(migrationsMap, singleSchemasDirs, types.MigrationTypeSingleSchema)
	dl.readMigrationsFromSchemaDirs(migrationsMap, tenantSchemasDirs, types.MigrationTypeTenantSchema)

	keys := make([]string, 0, len(migrationsMap))
	for key := range migrationsMap {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	var migrations []types.Migration
	for _, key := range keys {
		ms := migrationsMap[key]
		for _, m := range ms {
			migrations = append(migrations, m)
		}
	}

	return migrations
}

func (dl *DiskLoader) filterSchemaDirs(files []os.FileInfo, schemaDirs []string) []string {
	var dirs []string
	for _, f := range files {
		if f.IsDir() {
			name := f.Name()
			if utils.Contains(schemaDirs, &name) {
				dirs = append(dirs, name)
			}
		}
	}
	return dirs
}

func (dl *DiskLoader) readMigrationsFromSchemaDirs(migrations map[string][]types.Migration, sourceDirs []string, migrationType types.MigrationType) {
	for _, sourceDir := range sourceDirs {
		files, err := ioutil.ReadDir(filepath.Join(dl.Config.BaseDir, sourceDir))
		if err != nil {
			log.Panicf("Could not read migration source dir: %v", err)
		}
		for _, file := range files {
			if !file.IsDir() {
				mdef := types.MigrationDefinition{file.Name(), sourceDir, filepath.Join(sourceDir, file.Name()), migrationType}
				contents, err := ioutil.ReadFile(filepath.Join(dl.Config.BaseDir, mdef.File))
				if err != nil {
					log.Panicf("Could not read migration contents: %v", err)
				}
				m := types.Migration{mdef, string(contents)}
				e, ok := migrations[m.Name]
				if ok {
					e = append(e, m)
				} else {
					e = []types.Migration{m}
				}
				migrations[m.Name] = e
			}
		}
	}
}
