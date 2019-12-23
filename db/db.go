package db

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/lukaszbudnik/migrator/common"
	"github.com/lukaszbudnik/migrator/config"
	"github.com/lukaszbudnik/migrator/types"
)

// Connector interface abstracts all DB operations performed by migrator
type Connector interface {
	Init() error
	GetTenants() ([]string, error)
	GetDBMigrations() ([]types.MigrationDB, error)
	ApplyMigrations(context.Context, []types.Migration) (*types.MigrationResults, error)
	AddTenantAndApplyMigrations(context.Context, string, []types.Migration) (*types.MigrationResults, error)
	Dispose()
}

// baseConnector struct is a base struct for implementing DB specific dialects
type baseConnector struct {
	config  *config.Config
	dialect dialect
	db      *sql.DB
}

// NewConnector constructs Connector instance based on the passed Config
func NewConnector(config *config.Config) (Connector, error) {
	dialect, err := newDialect(config)
	if err != nil {
		return nil, err
	}
	connector := &baseConnector{config, dialect, nil}
	return connector, nil
}

const (
	migratorSchema           = "migrator"
	migratorTenantsTable     = "migrator_tenants"
	migratorMigrationsTable  = "migrator_migrations"
	defaultSchemaPlaceHolder = "{schema}"
)

// Init initialises connector by opening a connection to database
func (bc *baseConnector) Init() error {
	db, err := sql.Open(bc.config.Driver, bc.config.DataSource)
	if err != nil {
		return err
	}
	return bc.doInit(db)
}

// Init initialises connector by opening a connection to database
func (bc *baseConnector) doInit(db *sql.DB) error {
	if err := db.Ping(); err != nil {
		return fmt.Errorf("Failed to connect to database: %v", err)
	}
	bc.db = db

	tx, err := bc.db.Begin()
	if err != nil {
		return fmt.Errorf("Could not start DB transaction: %v", err)
	}

	// make sure migrator schema exists
	createSchema := bc.dialect.GetCreateSchemaSQL(migratorSchema)
	if _, err := bc.db.Query(createSchema); err != nil {
		return fmt.Errorf("Could not create migrator schema: %v", err)
	}

	// make sure migrations table exists
	createMigrationsTable := bc.dialect.GetCreateMigrationsTableSQL()
	if _, err := bc.db.Query(createMigrationsTable); err != nil {
		return fmt.Errorf("Could not create migrations table: %v", err)
	}

	// if using default migrator tenants table make sure it exists
	if bc.config.TenantSelectSQL == "" {
		createTenantsTable := bc.dialect.GetCreateTenantsTableSQL()
		if _, err := bc.db.Query(createTenantsTable); err != nil {
			return fmt.Errorf("Could not create default tenants table: %v", err)
		}
	}

	return tx.Commit()
}

// Dispose closes all resources allocated by connector
func (bc *baseConnector) Dispose() {
	if bc.db != nil {
		bc.db.Close()
	}
}

// getTenantSelectSQL returns SQL to be executed to list all DB tenants
func (bc *baseConnector) getTenantSelectSQL() string {
	var tenantSelectSQL string
	if bc.config.TenantSelectSQL != "" {
		tenantSelectSQL = bc.config.TenantSelectSQL
	} else {
		tenantSelectSQL = bc.dialect.GetTenantSelectSQL()
	}
	return tenantSelectSQL
}

// GetTenants returns a list of all DB tenants
func (bc *baseConnector) GetTenants() (tenants []string, err error) {
	tenantSelectSQL := bc.getTenantSelectSQL()

	tenants = []string{}

	rows, err := bc.db.Query(tenantSelectSQL)
	if err != nil {
		err = fmt.Errorf("Could not query tenants: %v", err)
		return
	}

	for rows.Next() {
		var name string
		if err = rows.Scan(&name); err != nil {
			err = fmt.Errorf("Could not read tenants: %v", err)
			return
		}
		tenants = append(tenants, name)
	}
	return
}

// GetDBMigrations returns a list of all applied DB migrations
func (bc *baseConnector) GetDBMigrations() (dbMigrations []types.MigrationDB, err error) {
	query := bc.dialect.GetMigrationSelectSQL()

	dbMigrations = []types.MigrationDB{}

	rows, err := bc.db.Query(query)
	if err != nil {
		err = fmt.Errorf("Could not query DB migrations: %v", err)
		return
	}

	for rows.Next() {
		var (
			name          string
			sourceDir     string
			filename      string
			migrationType types.MigrationType
			schema        string
			created       time.Time
			contents      string
			checksum      string
		)
		if err = rows.Scan(&name, &sourceDir, &filename, &migrationType, &schema, &created, &contents, &checksum); err != nil {
			err = fmt.Errorf("Could not read DB migration: %v", err)
			return
		}
		mdef := types.Migration{Name: name, SourceDir: sourceDir, File: filename, MigrationType: migrationType, Contents: contents, CheckSum: checksum}
		dbMigrations = append(dbMigrations, types.MigrationDB{Migration: mdef, Schema: schema, Created: created})
	}

	return
}

// ApplyMigrations applies passed migrations
func (bc *baseConnector) ApplyMigrations(ctx context.Context, migrations []types.Migration) (result *types.MigrationResults, err error) {
	if len(migrations) == 0 {
		return &types.MigrationResults{}, nil
	}

	tenants, err := bc.GetTenants()
	if err != nil {
		return nil, err
	}

	tx, err := bc.db.Begin()
	if err != nil {
		return nil, err
	}

	defer func() {
		r := recover()
		if r == nil {
			err = tx.Commit()
		} else {
			common.LogInfo(ctx, "Recovered in ApplyMigrations. Transaction rollback.")
			tx.Rollback()
			var ok bool
			err, ok = r.(error)
			if !ok {
				err = fmt.Errorf("%v", r)
			}
		}
	}()

	results := bc.applyMigrationsInTx(ctx, tx, tenants, migrations)
	return results, nil
}

// AddTenantAndApplyMigrations adds new tenant and applies all existing tenant migrations
func (bc *baseConnector) AddTenantAndApplyMigrations(ctx context.Context, tenant string, migrations []types.Migration) (results *types.MigrationResults, err error) {
	tenantInsertSQL := bc.getTenantInsertSQL()

	tx, err := bc.db.Begin()
	if err != nil {
		return
	}

	defer func() {
		r := recover()
		if r == nil {
			err = tx.Commit()
		} else {
			common.LogInfo(ctx, "Recovered in AddTenantAndApplyMigrations. Transaction rollback.")
			tx.Rollback()
			var ok bool
			err, ok = r.(error)
			if !ok {
				err = fmt.Errorf("%v", r)
			}
		}
	}()

	createSchema := bc.dialect.GetCreateSchemaSQL(tenant)
	if _, err = tx.Exec(createSchema); err != nil {
		common.LogPanic(ctx, "Create schema failed, transaction rollback was called: %v", err)
	}

	insert, err := bc.db.Prepare(tenantInsertSQL)
	if err != nil {
		common.LogPanic(ctx, "Could not create prepared statement: %v", err)
	}

	_, err = tx.Stmt(insert).Exec(tenant)
	if err != nil {
		common.LogPanic(ctx, "Failed to add tenant entry: %v", err)
	}

	results = bc.applyMigrationsInTx(ctx, tx, []string{tenant}, migrations)

	return
}

// getTenantInsertSQL returns tenant insert SQL statement from configuration file
// or, if absent, returns default Dialect-specific migrator tenant insert SQL
func (bc *baseConnector) getTenantInsertSQL() string {
	var tenantInsertSQL string
	// if set explicitly in config use it
	// otherwise use default value provided by Dialect implementation
	if bc.config.TenantInsertSQL != "" {
		tenantInsertSQL = bc.config.TenantInsertSQL
	} else {
		tenantInsertSQL = bc.dialect.GetTenantInsertSQL()
	}
	return tenantInsertSQL
}

// GetSchemaPlaceHolder returns a schema placeholder which is
// either the default one or overridden by user in config
func (bc *baseConnector) getSchemaPlaceHolder() string {
	var schemaPlaceHolder string
	if bc.config.SchemaPlaceHolder != "" {
		schemaPlaceHolder = bc.config.SchemaPlaceHolder
	} else {
		schemaPlaceHolder = defaultSchemaPlaceHolder
	}
	return schemaPlaceHolder
}

func (bc *baseConnector) applyMigrationsInTx(ctx context.Context, tx *sql.Tx, tenants []string, migrations []types.Migration) *types.MigrationResults {

	results := &types.MigrationResults{}
	results.StartedAt = time.Now()
	results.Tenants = len(tenants)

	defer func() {
		results.Duration = time.Now().Sub(results.StartedAt)
		results.MigrationsGrandTotal = results.TenantMigrationsTotal + results.SingleMigrations
		results.ScriptsGrandTotal = results.TenantScriptsTotal + results.SingleScripts
	}()

	schemaPlaceHolder := bc.getSchemaPlaceHolder()

	insertMigrationSQL := bc.dialect.GetMigrationInsertSQL()
	insert, err := bc.db.Prepare(insertMigrationSQL)
	if err != nil {
		common.LogPanic(ctx, "Could not create prepared statement: %v", err)
	}

	for _, m := range migrations {
		var schemas []string
		if m.MigrationType == types.MigrationTypeTenantMigration || m.MigrationType == types.MigrationTypeTenantScript {
			schemas = tenants
		} else {
			schemas = []string{m.SourceDir}
		}

		for _, s := range schemas {
			common.LogInfo(ctx, "Applying migration type: %d, schema: %s, file: %s ", m.MigrationType, s, m.File)

			contents := strings.Replace(m.Contents, schemaPlaceHolder, s, -1)

			_, err = tx.Exec(contents)
			if err != nil {
				common.LogPanic(ctx, "SQL migration failed: %v", err)
			}

			_, err = tx.Stmt(insert).Exec(m.Name, m.SourceDir, m.File, m.MigrationType, s, m.Contents, m.CheckSum)
			if err != nil {
				common.LogPanic(ctx, "Failed to add migration entry: %v", err)
			}
		}

		if m.MigrationType == types.MigrationTypeSingleMigration {
			results.SingleMigrations++
		}
		if m.MigrationType == types.MigrationTypeSingleScript {
			results.SingleScripts++
		}
		if m.MigrationType == types.MigrationTypeTenantMigration {
			results.TenantMigrations++
			results.TenantMigrationsTotal += len(schemas)
		}
		if m.MigrationType == types.MigrationTypeTenantScript {
			results.TenantScripts++
			results.TenantScriptsTotal += len(schemas)
		}

	}

	return results
}
