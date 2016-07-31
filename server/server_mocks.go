package server

import (
	"github.com/lukaszbudnik/migrator/config"
	"github.com/lukaszbudnik/migrator/db"
	"github.com/lukaszbudnik/migrator/loader"
	"github.com/lukaszbudnik/migrator/types"
	"time"
)

type mockedDiskLoader struct {
}

func (m *mockedDiskLoader) GetMigrations() []types.Migration {
	m1 := types.MigrationDefinition{"201602220000.sql", "source", "source/201602220000.sql", types.MigrationTypeSingleSchema}
	m2 := types.MigrationDefinition{"201602220001.sql", "source", "source/201602220001.sql", types.MigrationTypeSingleSchema}
	return []types.Migration{{m1, "select abc"}, {m2, "select def"}}
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

func (m *mockedConnector) GetTenants() []string {
	return []string{"a", "b", "c"}
}

func (m *mockedConnector) GetMigrations() []types.MigrationDB {
	m1 := types.MigrationDefinition{"201602220000.sql", "source", "source/201602220000.sql", types.MigrationTypeSingleSchema}
	d1 := time.Date(2016, 02, 22, 16, 41, 1, 123, time.UTC)
	ms := []types.MigrationDB{{m1, "source", d1}}

	return ms
}

func (m *mockedConnector) ApplyMigrations(migrations []types.Migration) {
}

func (m *mockedConnector) applyMigrationsWithInsertMigrationSQL(migrations []types.Migration, insertMigrationSQL string) {
}

func createMockedConnector(config *config.Config) db.Connector {
	return new(mockedConnector)
}
