package core

import (
	"github.com/lukaszbudnik/migrator/config"
	"github.com/lukaszbudnik/migrator/db"
	"github.com/lukaszbudnik/migrator/loader"
	"github.com/lukaszbudnik/migrator/types"
	"time"
)

type mockedDiskLoader struct {
}

func (m *mockedDiskLoader) GetDiskMigrations() []types.Migration {
	m1 := types.MigrationDefinition{Name: "201602220000.sql", SourceDir: "source", File: "source/201602220000.sql", MigrationType: types.MigrationTypeSingleSchema}
	m2 := types.MigrationDefinition{Name: "201602220001.sql", SourceDir: "source", File: "source/201602220001.sql", MigrationType: types.MigrationTypeSingleSchema}
	return []types.Migration{{MigrationDefinition: m1, Contents: "select abc"}, {MigrationDefinition: m2, Contents: "select def"}}
}

func createMockedDiskLoader(config *config.Config) loader.Loader {
	return new(mockedDiskLoader)
}

type mockedConnector struct {
}

func (m *mockedConnector) Init() {
}

func (m *mockedConnector) Dispose() {
}

func (m *mockedConnector) GetSchemaPlaceHolder() string {
	return ""
}

func (m *mockedConnector) GetTenantSelectSQL() string {
	return ""
}

func (m *mockedConnector) GetMigrationInsertSQL() string {
	return ""
}

func (m *mockedConnector) GetTenantInsertSQL() string {
	return ""
}

func (m *mockedConnector) GetTenants() []string {
	return []string{"a", "b", "c"}
}

func (m *mockedConnector) AddTenantAndApplyMigrations(string, []types.Migration) {
}

func (m *mockedConnector) GetDBMigrations() []types.MigrationDB {
	m1 := types.MigrationDefinition{Name: "201602220000.sql", SourceDir: "source", File: "source/201602220000.sql", MigrationType: types.MigrationTypeSingleSchema}
	d1 := time.Date(2016, 02, 22, 16, 41, 1, 123, time.UTC)
	ms := []types.MigrationDB{{MigrationDefinition: m1, Schema: "source", Created: d1}}

	return ms
}

func (m *mockedConnector) ApplyMigrations(migrations []types.Migration) {
}

func (m *mockedConnector) applyMigrationsWithInsertMigrationSQL(migrations []types.Migration, insertMigrationSQL string) {
}

func createMockedConnector(config *config.Config) db.Connector {
	return new(mockedConnector)
}
