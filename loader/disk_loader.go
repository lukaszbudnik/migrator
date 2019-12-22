package loader

import (
	"crypto/sha256"
	"encoding/hex"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"fmt"

	"github.com/lukaszbudnik/migrator/config"
	"github.com/lukaszbudnik/migrator/types"
)

// diskLoader is struct used for implementing Loader interface for loading migrations from disk
type diskLoader struct {
	config *config.Config
}

// GetDiskMigrations returns all migrations from disk
func (dl *diskLoader) GetDiskMigrations() (migrations []types.Migration, err error) {
	defer func() {
		if r := recover(); r != nil {
			var ok bool
			err, ok = r.(error)
			if !ok {
				err = fmt.Errorf("%v", r)
			}
		}
	}()

	migrations = []types.Migration{}

	absBaseDir, err := filepath.Abs(dl.config.BaseDir)
	if err != nil {
		panic(err.Error())
	}

	var dirs []string
	err = filepath.Walk(absBaseDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			dirs = append(dirs, path)
		}
		return nil
	})

	if err != nil {
		panic(err.Error())
	}

	singleMigrationsDirs := dl.filterSchemaDirs(dirs, dl.config.SingleMigrations)
	tenantMigrationsDirs := dl.filterSchemaDirs(dirs, dl.config.TenantMigrations)
	singleScriptsDirs := dl.filterSchemaDirs(dirs, dl.config.SingleScripts)
	tenantScriptsDirs := dl.filterSchemaDirs(dirs, dl.config.TenantScripts)

	migrationsMap := make(map[string][]types.Migration)

	dl.readFromDirs(migrationsMap, singleMigrationsDirs, types.MigrationTypeSingleMigration)
	dl.readFromDirs(migrationsMap, tenantMigrationsDirs, types.MigrationTypeTenantMigration)
	dl.readFromDirs(migrationsMap, singleScriptsDirs, types.MigrationTypeSingleScript)
	dl.readFromDirs(migrationsMap, tenantScriptsDirs, types.MigrationTypeTenantScript)

	keys := make([]string, 0, len(migrationsMap))
	for key := range migrationsMap {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	for _, key := range keys {
		ms := migrationsMap[key]
		for _, m := range ms {
			migrations = append(migrations, m)
		}
	}

	return
}

func (dl *diskLoader) filterSchemaDirs(dirs []string, migrationsDirs []string) []string {
	var filteredDirs []string
	for _, dir := range dirs {
		for _, migrationsDir := range migrationsDirs {
			if strings.HasSuffix(dir, migrationsDir) {
				filteredDirs = append(filteredDirs, dir)
			}
		}
	}
	return filteredDirs
}

func (dl *diskLoader) readFromDirs(migrations map[string][]types.Migration, sourceDirs []string, migrationType types.MigrationType) {
	for _, sourceDir := range sourceDirs {
		files, err := ioutil.ReadDir(sourceDir)
		if err != nil {
			panic(err.Error())
		}
		for _, file := range files {
			if !file.IsDir() {
				contents, err := ioutil.ReadFile(filepath.Join(sourceDir, file.Name()))
				if err != nil {
					panic(err.Error())
				}
				hasher := sha256.New()
				hasher.Write([]byte(contents))
				name := strings.Replace(file.Name(), dl.config.BaseDir, "", 1)
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
