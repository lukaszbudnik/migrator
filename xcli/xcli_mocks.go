package xcli

import (
	"github.com/lukaszbudnik/migrator/config"
	"github.com/lukaszbudnik/migrator/db"
	"github.com/lukaszbudnik/migrator/loader"
	"github.com/lukaszbudnik/migrator/types"
)

type mockedDiskLoader struct {
}

func (m *mockedDiskLoader) GetMigrations() []types.Migration {
	// returns empty array
	return []types.Migration{}
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
	// returns empty array
	return []string{}
}

func (m *mockedConnector) GetMigrations() []types.MigrationDB {
	// returns empty array
	return []types.MigrationDB{}
}

func (m *mockedConnector) ApplyMigrations(migrations []types.Migration) {
}

func (m *mockedConnector) applyMigrationsWithInsertMigrationSQL(migrations []types.Migration, insertMigrationSQL string) {
}

func createMockedConnector(config *config.Config) db.Connector {
	return new(mockedConnector)
}
