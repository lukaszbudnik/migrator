package main

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
)

func filterSchemaDirs(sourceDir string, files []os.FileInfo, schemaDirs []string) []string {
	var dirs []string
	for _, f := range files {
		if f.IsDir() {
			if stringInSlice(f.Name(), schemaDirs) {
				dirs = append(dirs, filepath.Join(sourceDir, f.Name()))
			}
		}
	}
	return dirs
}

func readMigrations(migrations map[string][]string, sourceDirs []string) error {
	for _, f := range sourceDirs {
		ms, err := ioutil.ReadDir(f)
		if err != nil {
			return err
		}
		for _, m := range ms {
			if !m.IsDir() {
				e, ok := migrations[m.Name()]
				if ok {
					e = append(e, filepath.Join(f, m.Name()))
				} else {
					e = []string{filepath.Join(f, m.Name())}
				}
				migrations[m.Name()] = e
			}
		}
	}
	return nil
}

func listAllMigrations(config Config) ([]string, error) {

	sourceDir, err := filepath.Rel(".", config.SourceDir)
	if err != nil {
		return nil, err
	}

	dirs, err := ioutil.ReadDir(sourceDir)
	if err != nil {
		return nil, err
	}

	singleSchemasDirs := filterSchemaDirs(sourceDir, dirs, config.SingleSchemas)
	tenantSchemasDirs := filterSchemaDirs(sourceDir, dirs, config.TenantSchemas)

	migrationsMap := make(map[string][]string)

	readMigrations(migrationsMap, singleSchemasDirs)
	readMigrations(migrationsMap, tenantSchemasDirs)

	keys := make([]string, 0, len(migrationsMap))
	for key := range migrationsMap {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	var migrations []string
	for _, key := range keys {
		ms := migrationsMap[key]
		for _, m := range ms {
			migrations = append(migrations, m)
		}
	}

	return migrations, nil

}

func computeMigrationsToApply(allMigrations []string, dbMigrations []string) []string {
	var (
		lenMin  int
		longest []string
		out     []string
	)
	if len(allMigrations) < len(dbMigrations) {
		lenMin = len(allMigrations)
		longest = dbMigrations
	} else {
		lenMin = len(dbMigrations)
		longest = allMigrations
	}
	for i := 0; i < lenMin; i++ {
		if allMigrations[i] != dbMigrations[i] {
			out = append(out, allMigrations[i])
			out = append(out, dbMigrations[i])
		}
	}
	for _, v := range longest[lenMin:] {
		out = append(out, v)
	}
	return out
}
