package loader

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/lukaszbudnik/migrator/types"
)

// diskLoader is struct used for implementing Loader interface for loading migrations from disk
type diskLoader struct {
	baseLoader
}

// GetSourceMigrations returns all migrations from disk
func (dl *diskLoader) GetSourceMigrations() []types.Migration {
	migrations := []types.Migration{}

	absBaseDir, err := filepath.Abs(dl.config.BaseLocation)
	if err != nil {
		panic(fmt.Sprintf("Could not convert baseLocation to absolute path: %v", err.Error()))
	}

	singleMigrationsDirs := dl.getDirs(absBaseDir, dl.config.SingleMigrations)
	tenantMigrationsDirs := dl.getDirs(absBaseDir, dl.config.TenantMigrations)
	singleScriptsDirs := dl.getDirs(absBaseDir, dl.config.SingleScripts)
	tenantScriptsDirs := dl.getDirs(absBaseDir, dl.config.TenantScripts)

	migrationsMap := make(map[string][]types.Migration)
	dl.readFromDirs(migrationsMap, singleMigrationsDirs, types.MigrationTypeSingleMigration)
	dl.readFromDirs(migrationsMap, tenantMigrationsDirs, types.MigrationTypeTenantMigration)
	dl.sortMigrations(migrationsMap, &migrations)

	migrationsMap = make(map[string][]types.Migration)
	dl.readFromDirs(migrationsMap, singleScriptsDirs, types.MigrationTypeSingleScript)
	dl.sortMigrations(migrationsMap, &migrations)

	migrationsMap = make(map[string][]types.Migration)
	dl.readFromDirs(migrationsMap, tenantScriptsDirs, types.MigrationTypeTenantScript)
	dl.sortMigrations(migrationsMap, &migrations)

	return migrations
}

func (dl *diskLoader) getDirs(baseDir string, migrationsDirs []string) []string {
	var filteredDirs []string
	for _, migrationsDir := range migrationsDirs {
		filteredDirs = append(filteredDirs, filepath.Join(baseDir, migrationsDir))
	}
	return filteredDirs
}

func (dl *diskLoader) readFromDirs(migrations map[string][]types.Migration, sourceDirs []string, migrationType types.MigrationType) {
	for _, sourceDir := range sourceDirs {
		files, err := ioutil.ReadDir(sourceDir)
		if err != nil {
			panic(fmt.Sprintf("Could not read source dir %v: %v", sourceDir, err.Error()))
		}
		for _, file := range files {
			if !file.IsDir() {
				fullPath := filepath.Join(sourceDir, file.Name())
				contents, err := ioutil.ReadFile(fullPath)
				if err != nil {
					panic(fmt.Sprintf("Could not read file %v: %v", fullPath, err.Error()))
				}
				hasher := sha256.New()
				hasher.Write([]byte(contents))
				name := strings.Replace(file.Name(), dl.config.BaseLocation, "", 1)
				m := types.Migration{Name: name, SourceDir: sourceDir, File: filepath.Join(sourceDir, file.Name()), MigrationType: migrationType, Contents: string(contents), CheckSum: hex.EncodeToString(hasher.Sum(nil))}

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
