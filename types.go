package main

import (
	"bytes"
	"fmt"
	"strings"
	"time"
)

// MigrationType stores information about type of migration
type MigrationType uint32

const (
	// ModeSingleSchema is used to mark single migration
	ModeSingleSchema MigrationType = 1
	// ModeTenantSchema is used to mark tenant migrations
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
}

// DBMigration embeds MigrationDefinition and contain other DB properties
type DBMigration struct {
	MigrationDefinition
	Schema  string
	Created time.Time
}

func (m Migration) String() string {
	return fmt.Sprintf("| %-10s | %-20s | %-30s | %4d |", m.SourceDir, m.Name, m.File, m.MigrationType)
}

func (m DBMigration) String() string {
	created := fmt.Sprintf("%v", m.Created)
	index := strings.Index(created, ".")
	created = created[:index]
	return fmt.Sprintf("| %-10s | %-20s | %-30s | %-10s | %-20s | %4d |", m.SourceDir, m.Name, m.File, m.Schema, created, m.MigrationType)
}

func migrationsString(migrations []Migration) string {
	var buffer bytes.Buffer

	buffer.WriteString("+")
	buffer.WriteString(strings.Repeat("-", 75))
	buffer.WriteString("+\n")

	buffer.WriteString(fmt.Sprintf("| %-10s | %-20s | %-30s | %4s |\n", "SourceDir", "Name", "File", "Type"))

	buffer.WriteString("+")
	buffer.WriteString(strings.Repeat("-", 75))
	buffer.WriteString("+\n")

	for _, m := range migrations {
		buffer.WriteString(fmt.Sprintf("%v\n", m))
	}

	buffer.WriteString("+")
	buffer.WriteString(strings.Repeat("-", 75))
	buffer.WriteString("+")

	return buffer.String()
}

func dbMigrationsString(migrations []DBMigration) string {
	var buffer bytes.Buffer

	buffer.WriteString("+")
	buffer.WriteString(strings.Repeat("-", 111))
	buffer.WriteString("+\n")

	buffer.WriteString(fmt.Sprintf("| %-10s | %-20s | %-30s | %-10s | %-20s | %4s |\n", "SourceDir", "Name", "File", "Schema", "Created", "Type"))

	buffer.WriteString("+")
	buffer.WriteString(strings.Repeat("-", 111))
	buffer.WriteString("+\n")

	for _, m := range migrations {
		buffer.WriteString(fmt.Sprintf("%v\n", m))
	}

	buffer.WriteString("+")
	buffer.WriteString(strings.Repeat("-", 111))
	buffer.WriteString("+")

	return buffer.String()
}
