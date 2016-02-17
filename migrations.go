package main

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	// "time"
)

// MigrationType stores information about type of migration
type MigrationType uint32

const (
	// ModeSingleSchema is used to mark single schema migrations
	ModeSingleSchema MigrationType = 1
	// ModeTenantSchema is used to mark tenant schema migrations
	ModeTenantSchema MigrationType = 2
)

// MigrationDefinition contains basic information about migration
type MigrationDefinition struct {
	Name          string
	SourceDir     string
	File          string
	MigrationType MigrationType
}

// Migration embeds MigrationDefinition and contains its contents
type Migration struct {
	MigrationDefinition
	Contents string
	// Schema    string
	// Created		time.Time
}

func filterSchemaDirs(sourceDir string, files []os.FileInfo, schemaDirs []string) []string {
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

func readMigrations(migrations map[string][]MigrationDefinition, baseDir string, sourceDirs []string, migrationType MigrationType) error {
	for _, f := range sourceDirs {
		ms, err := ioutil.ReadDir(filepath.Join(baseDir, f))
		if err != nil {
			return err
		}
		for _, m := range ms {
			if !m.IsDir() {
				e, ok := migrations[m.Name()]
				migration := MigrationDefinition{m.Name(), f, filepath.Join(f, m.Name()), migrationType}
				if ok {
					e = append(e, migration)
				} else {
					e = []MigrationDefinition{migration}
				}
				migrations[m.Name()] = e
			}
		}
	}
	return nil
}

func listAllMigrations(config Config) ([]MigrationDefinition, error) {

	baseDir, err := filepath.Rel(".", config.BaseDir)
	if err != nil {
		return nil, err
	}

	dirs, err := ioutil.ReadDir(baseDir)
	if err != nil {
		return nil, err
	}

	singleSchemasDirs := filterSchemaDirs(baseDir, dirs, config.SingleSchemas)
	tenantSchemasDirs := filterSchemaDirs(baseDir, dirs, config.TenantSchemas)

	migrationsMap := make(map[string][]MigrationDefinition)

	readMigrations(migrationsMap, baseDir, singleSchemasDirs, ModeSingleSchema)
	readMigrations(migrationsMap, baseDir, tenantSchemasDirs, ModeTenantSchema)

	keys := make([]string, 0, len(migrationsMap))
	for key := range migrationsMap {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	var migrations []MigrationDefinition
	for _, key := range keys {
		ms := migrationsMap[key]
		for _, m := range ms {
			migrations = append(migrations, m)
		}
	}

	return migrations, nil

}

func computeMigrationsToApply(allMigrations []MigrationDefinition, dbMigrations []MigrationDefinition) []MigrationDefinition {
	var (
		lenMin  int
		longest []MigrationDefinition
		out     []MigrationDefinition
	)
	if len(allMigrations) < len(dbMigrations) {
		lenMin = len(allMigrations)
		longest = dbMigrations
	} else {
		lenMin = len(dbMigrations)
		longest = allMigrations
	}
	for i := 0; i < lenMin; i++ {
		if (allMigrations[i].Name != dbMigrations[i].Name) &&
			(allMigrations[i].SourceDir != dbMigrations[i].SourceDir) &&
			(allMigrations[i].File != dbMigrations[i].File) {
			out = append(out, allMigrations[i])
		}
	}
	for _, v := range longest[lenMin:] {
		out = append(out, v)
	}
	return out
}

func loadMigrations(config Config, migrationsDef []MigrationDefinition) ([]Migration, error) {
	migrations := make([]Migration, 0, len(migrationsDef))

	for _, mdef := range migrationsDef {
		contents, err := ioutil.ReadFile(filepath.Join(config.BaseDir, mdef.SourceDir, mdef.Name))
		if err != nil {
			return nil, err
		}
		m := Migration{mdef, string(contents)}
		migrations = append(migrations, m)
	}

	return migrations, nil
}
