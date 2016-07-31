package types

import (
	"bytes"
	"fmt"
	"strings"
	"time"
)

// MigrationType stores information about type of migration
type MigrationType uint32

const (
	// MigrationTypeSingleSchema is used to mark single migration
	MigrationTypeSingleSchema MigrationType = 1
	// MigrationTypeTenantSchema is used to mark tenant migrations
	MigrationTypeTenantSchema MigrationType = 2
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

// MigrationDB embeds MigrationDefinition and contain other DB properties
type MigrationDB struct {
	MigrationDefinition
	Schema  string
	Created time.Time
}

func (m Migration) String() string {
	return fmt.Sprintf("| %-10s | %-20s | %-30s | %4d |", m.SourceDir, m.Name, m.File, m.MigrationType)
}

func (m MigrationDB) String() string {
	created := fmt.Sprintf("%v", m.Created)
	index := strings.Index(created, ".")
	created = created[:index]
	return fmt.Sprintf("| %-10s | %-20s | %-30s | %-10s | %-20s | %4d |", m.SourceDir, m.Name, m.File, m.Schema, created, m.MigrationType)
}

// MigrationArrayString creates a string representation of migrations list
func MigrationArrayString(migrations []Migration) string {
	var buffer bytes.Buffer

	buffer.WriteString("+")
	buffer.WriteString(strings.Repeat("-", 75))
	buffer.WriteString("+\n")

	buffer.WriteString(fmt.Sprintf("| %-10s | %-20s | %-30s | %4s |\n", "SourceDir", "Name", "File", "Type"))

	buffer.WriteString("+")
	buffer.WriteString(strings.Repeat("-", 75))
	buffer.WriteString("+\n")

	if len(migrations) > 0 {
		for _, m := range migrations {
			buffer.WriteString(fmt.Sprintf("%v\n", m))
		}
	} else {
		buffer.WriteString(fmt.Sprintf("| %-73s |\n", "Empty"))
	}

	buffer.WriteString("+")
	buffer.WriteString(strings.Repeat("-", 75))
	buffer.WriteString("+")

	return buffer.String()
}

// MigrationDBArrayString creates a string representation of DB migrations list
func MigrationDBArrayString(migrations []MigrationDB) string {
	var buffer bytes.Buffer

	buffer.WriteString("+")
	buffer.WriteString(strings.Repeat("-", 111))
	buffer.WriteString("+\n")

	buffer.WriteString(fmt.Sprintf("| %-10s | %-20s | %-30s | %-10s | %-20s | %4s |\n", "SourceDir", "Name", "File", "Schema", "Created", "Type"))

	buffer.WriteString("+")
	buffer.WriteString(strings.Repeat("-", 111))
	buffer.WriteString("+\n")

	if len(migrations) > 0 {
		for _, m := range migrations {
			buffer.WriteString(fmt.Sprintf("%v\n", m))
		}
	} else {
		buffer.WriteString(fmt.Sprintf("| %-109s |\n", "Empty"))
	}

	buffer.WriteString("+")
	buffer.WriteString(strings.Repeat("-", 111))
	buffer.WriteString("+")

	return buffer.String()
}

// TenantArrayString creates a string representation of DB tenants list
func TenantArrayString(dbTenants []string) string {
	var buffer bytes.Buffer

	buffer.WriteString("+")
	buffer.WriteString(strings.Repeat("-", 30))
	buffer.WriteString("+\n")

	for _, t := range dbTenants {
		buffer.WriteString(fmt.Sprintf("| %-28s |\n", t))
	}

	buffer.WriteString("+")
	buffer.WriteString(strings.Repeat("-", 30))
	buffer.WriteString("+")

	return buffer.String()

}
