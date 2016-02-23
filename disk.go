package main

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sort"
)

// Loader interface abstracts all disk operations performed by migrator
type Loader interface {
	GetDiskMigrations() []Migration
}

// DiskLoader struct is a base struct for implementing Loader interface
type DiskLoader struct {
	Config *Config
}

// CreateLoader abstracts all disk operations performed by migrator
func CreateLoader(config *Config) Loader {
	return &DiskLoader{config}
}

// GetDiskMigrations loads all migrations from disk
func (dl *DiskLoader) GetDiskMigrations() []Migration {
	dirs, err := ioutil.ReadDir(dl.Config.BaseDir)
	if err != nil {
		log.Panicf("Could not read migration base dir ==> %v", err)
	}

	singleSchemasDirs := dl.filterSchemaDirs(dirs, dl.Config.SingleSchemas)
	tenantSchemasDirs := dl.filterSchemaDirs(dirs, dl.Config.TenantSchemas)

	migrationsMap := make(map[string][]Migration)

	dl.readMigrationsFromSchemaDirs(migrationsMap, singleSchemasDirs, ModeSingleSchema)
	dl.readMigrationsFromSchemaDirs(migrationsMap, tenantSchemasDirs, ModeTenantSchema)

	keys := make([]string, 0, len(migrationsMap))
	for key := range migrationsMap {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	var migrations []Migration
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
			if stringInSlice(f.Name(), schemaDirs) {
				dirs = append(dirs, f.Name())
			}
		}
	}
	return dirs
}

func (dl *DiskLoader) readMigrationsFromSchemaDirs(migrations map[string][]Migration, sourceDirs []string, migrationType MigrationType) {
	for _, sourceDir := range sourceDirs {
		files, err := ioutil.ReadDir(filepath.Join(dl.Config.BaseDir, sourceDir))
		if err != nil {
			log.Panicf("Could not read migration source dir ==> %v", err)
		}
		for _, file := range files {
			if !file.IsDir() {
				mdef := MigrationDefinition{file.Name(), sourceDir, filepath.Join(sourceDir, file.Name()), migrationType}
				contents, err := ioutil.ReadFile(filepath.Join(dl.Config.BaseDir, mdef.File))
				if err != nil {
					log.Panicf("Could not read migration contents ==> %v", err)
				}
				m := Migration{mdef, string(contents)}
				e, ok := migrations[m.Name]
				if ok {
					e = append(e, m)
				} else {
					e = []Migration{m}
				}
				migrations[m.Name] = e
			}
		}
	}
}
