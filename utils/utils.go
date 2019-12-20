package utils

import (
	"bytes"
	"fmt"
	"io"
	"text/tabwriter"

	"github.com/lukaszbudnik/migrator/types"
)

// MigrationArrayToString creates a string representation of Migration array
func MigrationArrayToString(migrations []types.Migration) string {
	buffer := new(bytes.Buffer)
	w := tabwriter.NewWriter(buffer, 0, 0, 1, ' ', tabwriter.Debug)

	fmt.Fprintf(w, "%v \t %v \t %v \t %v \t %v", "SourceDir", "Name", "File", "Type", "CheckSum")

	for _, m := range migrations {
		formatMigration(w, &m)
	}

	w.Flush()
	return buffer.String()
}

func formatMigration(w io.Writer, m *types.Migration) {
	fmt.Fprintf(w, "\n%v \t %v \t %v \t %v \t %v", m.SourceDir, m.Name, m.File, m.MigrationType, m.CheckSum)
}

// MigrationDBArrayToString creates a string representation of MigrationDB array
func MigrationDBArrayToString(migrations []types.MigrationDB) string {
	buffer := new(bytes.Buffer)
	w := tabwriter.NewWriter(buffer, 0, 0, 1, ' ', tabwriter.Debug)

	fmt.Fprintf(w, "%v \t %v \t %v \t %v \t %v \t %v \t %v", "SourceDir", "Name", "File", "Schema", "Created", "Type", "CheckSum")

	for _, m := range migrations {
		formatMigrationDB(w, &m)
	}

	w.Flush()
	return buffer.String()
}

func formatMigrationDB(w io.Writer, m *types.MigrationDB) {
	fmt.Fprintf(w, "\n%v \t %v \t %v \t %v \t %v \t %v \t %v", m.SourceDir, m.Name, m.File, m.Schema, m.Created, m.MigrationType, m.CheckSum)
}

// TenantArrayToString creates a string representation of Tenant array
func TenantArrayToString(dbTenants []string) string {
	var buffer bytes.Buffer

	buffer.WriteString("Name")

	for _, t := range dbTenants {
		buffer.WriteString("\n")
		buffer.WriteString(t)
	}

	return buffer.String()

}
