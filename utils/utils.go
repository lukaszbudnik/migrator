package utils

import (
	"bytes"
	"fmt"
	"github.com/lukaszbudnik/migrator/types"
	"strings"
)

// Contains returns true when element is present in slice
func Contains(slice []string, element *string) bool {
	for _, a := range slice {
		if a == *element {
			return true
		}
	}
	return false
}

// MigrationToString creates a string representation of Migration
func MigrationToString(m *types.Migration) string {
	return fmt.Sprintf("| %-10s | %-20s | %-30s | %4d |", m.SourceDir, m.Name, m.File, m.MigrationType)
}

// MigrationDBToString creates a string representation of MigrationDB
func MigrationDBToString(m *types.MigrationDB) string {
	created := fmt.Sprintf("%v", m.Created)
	index := strings.Index(created, ".")
	created = created[:index]
	return fmt.Sprintf("| %-10s | %-20s | %-30s | %-10s | %-20s | %4d |", m.SourceDir, m.Name, m.File, m.Schema, created, m.MigrationType)
}

// MigrationArrayToString creates a string representation of Migration array
func MigrationArrayToString(migrations []types.Migration) string {
	var buffer bytes.Buffer

	buffer.WriteString("+")
	buffer.WriteString(strings.Repeat("-", 75))
	buffer.WriteString("+\n")

	buffer.WriteString(fmt.Sprintf("| %-10s | %-20s | %-30s | %4s |\n", "SourceDir", "Name", "File", "Type"))

	buffer.WriteString("+")
	buffer.WriteString(strings.Repeat("-", 75))
	buffer.WriteString("+\n")

	for _, m := range migrations {
		buffer.WriteString(fmt.Sprintf("%v\n", MigrationToString(&m)))
	}

	buffer.WriteString("+")
	buffer.WriteString(strings.Repeat("-", 75))
	buffer.WriteString("+")

	return buffer.String()
}

// MigrationDBArrayToString creates a string representation of MigrationDB array
func MigrationDBArrayToString(migrations []types.MigrationDB) string {
	var buffer bytes.Buffer

	buffer.WriteString("+")
	buffer.WriteString(strings.Repeat("-", 111))
	buffer.WriteString("+\n")

	buffer.WriteString(fmt.Sprintf("| %-10s | %-20s | %-30s | %-10s | %-20s | %4s |\n", "SourceDir", "Name", "File", "Schema", "Created", "Type"))

	buffer.WriteString("+")
	buffer.WriteString(strings.Repeat("-", 111))
	buffer.WriteString("+\n")

	for _, m := range migrations {
		buffer.WriteString(fmt.Sprintf("%v\n", MigrationDBToString(&m)))
	}

	buffer.WriteString("+")
	buffer.WriteString(strings.Repeat("-", 111))
	buffer.WriteString("+")

	return buffer.String()
}

// TenantArrayToString creates a string representation of Tenant array
func TenantArrayToString(dbTenants []string) string {
	var buffer bytes.Buffer

	buffer.WriteString("+")
	buffer.WriteString(strings.Repeat("-", 30))
	buffer.WriteString("+\n")

	buffer.WriteString(fmt.Sprintf("| %-28s |\n", "Name"))

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
