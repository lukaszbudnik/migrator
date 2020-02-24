package db

import (
	"context"
	"database/sql"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/graph-gophers/graphql-go"

	"github.com/lukaszbudnik/migrator/common"
	"github.com/lukaszbudnik/migrator/config"
	"github.com/lukaszbudnik/migrator/types"
)

// Connector interface abstracts all DB operations performed by migrator
type Connector interface {
	GetTenants() []types.Tenant
	GetVersions() []types.Version
	GetVersionsByFile(file string) []types.Version
	GetAppliedMigrations() []types.MigrationDB
	ApplyMigrations(types.MigrationsModeType, []types.Migration) *types.MigrationResults
	AddTenantAndApplyMigrations(types.MigrationsModeType, string, []types.Migration) *types.MigrationResults
	Dispose()
}

// baseConnector struct is a base struct for implementing DB specific dialects
type baseConnector struct {
	ctx     context.Context
	config  *config.Config
	dialect dialect
	db      *sql.DB
}

// Factory is a factory method for creating Loader instance
type Factory func(context.Context, *config.Config) Connector

// New constructs Connector instance based on the passed Config
func New(ctx context.Context, config *config.Config) Connector {
	dialect := newDialect(config)
	connector := &baseConnector{ctx, config, dialect, nil}
	connector.connect()
	connector.init()
	return connector
}

const (
	migratorSchema           = "migrator"
	migratorTenantsTable     = "migrator_tenants"
	migratorMigrationsTable  = "migrator_migrations"
	migratorVersionsTable    = "migrator_versions"
	defaultSchemaPlaceHolder = "{schema}"
)

// connect connects to a database
func (bc *baseConnector) connect() {
	db, err := sql.Open(bc.config.Driver, bc.config.DataSource)
	if err != nil {
		panic(fmt.Sprintf("Failed to open connction to DB: %v", err.Error()))
	}
	bc.db = db
}

// init initialises migrator by making sure proper schema/table are created
func (bc *baseConnector) init() {
	if err := bc.db.Ping(); err != nil {
		panic(fmt.Sprintf("Failed to connect to database: %v", err))
	}

	tx, err := bc.db.Begin()
	if err != nil {
		panic(fmt.Sprintf("Could not start DB transaction: %v", err))
	}

	// make sure migrator schema exists
	createSchema := bc.dialect.GetCreateSchemaSQL(migratorSchema)
	if _, err := bc.db.Query(createSchema); err != nil {
		panic(fmt.Sprintf("Could not create migrator schema: %v", err))
	}

	// make sure migrations table exists
	createMigrationsTable := bc.dialect.GetCreateMigrationsTableSQL()
	if _, err := bc.db.Query(createMigrationsTable); err != nil {
		panic(fmt.Sprintf("Could not create migrations table: %v", err))
	}

	// make sure versions table exists
	createVersionsTableSQLs := bc.dialect.GetCreateVersionsTableSQL()
	for _, createVersionsTableSQL := range createVersionsTableSQLs {
		if _, err := bc.db.Query(createVersionsTableSQL); err != nil {
			panic(fmt.Sprintf("Could not create versions table: %v", err))
		}
	}

	// if using default migrator tenants table make sure it exists
	if bc.config.TenantSelectSQL == "" {
		createTenantsTable := bc.dialect.GetCreateTenantsTableSQL()
		if _, err := bc.db.Query(createTenantsTable); err != nil {
			panic(fmt.Sprintf("Could not create default tenants table: %v", err))
		}
	}

	if err := tx.Commit(); err != nil {
		panic(fmt.Sprintf("Could not commit transaction: %v", err))
	}
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
func (bc *baseConnector) GetTenants() []types.Tenant {
	tenantSelectSQL := bc.getTenantSelectSQL()

	tenants := []types.Tenant{}

	rows, err := bc.db.Query(tenantSelectSQL)
	if err != nil {
		panic(fmt.Sprintf("Could not query tenants: %v", err))
	}

	for rows.Next() {
		var name string
		if err = rows.Scan(&name); err != nil {
			panic(fmt.Sprintf("Could not read tenants: %v", err))
		}
		tenants = append(tenants, types.Tenant{Name: name})
	}

	return tenants
}

func (bc *baseConnector) GetVersions() []types.Version {
	versionsSelectSQL := bc.dialect.GetVersionsSelectSQL()

	rows, err := bc.db.Query(versionsSelectSQL)
	if err != nil {
		panic(fmt.Sprintf("Could not query versions: %v", err))
	}

	return bc.readVersions(rows)
}

func (bc *baseConnector) GetVersionsByFile(file string) []types.Version {
	versionsSelectSQL := bc.dialect.GetVersionsByFileSQL()

	rows, err := bc.db.Query(versionsSelectSQL, file)
	if err != nil {
		panic(fmt.Sprintf("Could not query versions: %v", err))
	}

	return bc.readVersions(rows)
}

func (bc *baseConnector) readVersions(rows *sql.Rows) []types.Version {
	versions := []types.Version{}

	for rows.Next() {
		var id int64
		var name string
		var created time.Time
		if err := rows.Scan(&id, &name, &created); err != nil {
			panic(fmt.Sprintf("Could not read versions: %v", err))
		}
		versions = append(versions, types.Version{ID: int32(id), Name: name, Created: graphql.Time{Time: created}})
	}

	return versions
}

// GetAppliedMigrations returns a list of all applied DB migrations
func (bc *baseConnector) GetAppliedMigrations() []types.MigrationDB {
	query := bc.dialect.GetMigrationSelectSQL()

	dbMigrations := []types.MigrationDB{}

	rows, err := bc.db.Query(query)
	if err != nil {
		panic(fmt.Sprintf("Could not query DB migrations: %v", err.Error()))
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
			panic(fmt.Sprintf("Could not read DB migration: %v", err.Error()))
		}
		mdef := types.Migration{Name: name, SourceDir: sourceDir, File: filename, MigrationType: migrationType, Contents: contents, CheckSum: checksum}
		dbMigrations = append(dbMigrations, types.MigrationDB{Migration: mdef, Schema: schema, AppliedAt: graphql.Time{Time: created}, Created: graphql.Time{Time: created}})
	}
	return dbMigrations
}

// ApplyMigrations applies passed migrations
func (bc *baseConnector) ApplyMigrations(mode types.MigrationsModeType, migrations []types.Migration) *types.MigrationResults {
	if len(migrations) == 0 {
		return &types.MigrationResults{
			StartedAt: time.Now(),
			Duration:  0,
		}
	}

	tenants := bc.GetTenants()

	tx, err := bc.db.Begin()
	if err != nil {
		panic(fmt.Sprintf("Could not start transaction: %v", err.Error()))
	}

	defer func() {
		r := recover()
		if r == nil {
			if mode == types.ModeTypeDryRun {
				common.LogInfo(bc.ctx, "Running in dry-run mode, calling rollback")
				tx.Rollback()
			} else {
				common.LogInfo(bc.ctx, "Running in %v mode, committing transaction", mode)
				if err := tx.Commit(); err != nil {
					panic(fmt.Sprintf("Could not commit transaction: %v", err.Error()))
				}
			}
		} else {
			common.LogInfo(bc.ctx, "Recovered in ApplyMigrations. Transaction rollback.")
			tx.Rollback()
			panic(r)
		}
	}()

	return bc.applyMigrationsInTx(tx, mode, tenants, migrations)
}

// AddTenantAndApplyMigrations adds new tenant and applies all existing tenant migrations
func (bc *baseConnector) AddTenantAndApplyMigrations(mode types.MigrationsModeType, tenant string, migrations []types.Migration) *types.MigrationResults {
	tenantInsertSQL := bc.getTenantInsertSQL()

	tx, err := bc.db.Begin()
	if err != nil {
		panic(fmt.Sprintf("Could not start transaction: %v", err.Error()))
	}

	defer func() {
		r := recover()
		if r == nil {
			if mode == types.ModeTypeDryRun {
				common.LogInfo(bc.ctx, "Running in dry-run mode, calling rollback")
				tx.Rollback()
			} else {
				common.LogInfo(bc.ctx, "Running in %v mode, committing transaction", mode)
				if err := tx.Commit(); err != nil {
					panic(fmt.Sprintf("Could not commit transaction: %v", err.Error()))
				}
			}
		} else {
			common.LogInfo(bc.ctx, "Recovered in AddTenantAndApplyMigrations. Transaction rollback.")
			tx.Rollback()
			panic(r)
		}
	}()

	createSchema := bc.dialect.GetCreateSchemaSQL(tenant)
	if _, err = tx.Exec(createSchema); err != nil {
		panic(fmt.Sprintf("Create schema failed: %v", err))
	}

	insert, err := bc.db.Prepare(tenantInsertSQL)
	if err != nil {
		panic(fmt.Sprintf("Could not create prepared statement: %v", err))
	}

	_, err = tx.Stmt(insert).Exec(tenant)
	if err != nil {
		panic(fmt.Sprintf("Failed to add tenant entry: %v", err))
	}

	tenantStruct := types.Tenant{Name: tenant}
	results := bc.applyMigrationsInTx(tx, mode, []types.Tenant{tenantStruct}, migrations)

	return results
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

// getSchemaPlaceHolder returns a schema placeholder which is
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

func (bc *baseConnector) applyMigrationsInTx(tx *sql.Tx, mode types.MigrationsModeType, tenants []types.Tenant, migrations []types.Migration) *types.MigrationResults {

	results := &types.MigrationResults{
		StartedAt: time.Now(),
		Tenants:   len(tenants),
	}

	defer func() {
		results.Duration = time.Now().Sub(results.StartedAt)
		results.MigrationsGrandTotal = results.TenantMigrationsTotal + results.SingleMigrations
		results.ScriptsGrandTotal = results.TenantScriptsTotal + results.SingleScripts
	}()

	schemaPlaceHolder := bc.getSchemaPlaceHolder()

	var versionID int64
	versionInsertSQL := bc.dialect.GetVersionInsertSQL()
	versionInsert, err := bc.db.Prepare(versionInsertSQL)
	if err != nil {
		panic(fmt.Sprintf("Could not create prepared statement for version: %v", err))
	}
	stmt := tx.Stmt(versionInsert)
	if bc.dialect.LastInsertIDSupported() {
		result, _ := stmt.Exec("")
		versionID, _ = result.LastInsertId()
	} else {
		stmt.QueryRow("").Scan(&versionID)
	}

	insertMigrationSQL := bc.dialect.GetMigrationInsertSQL()
	insert, err := bc.db.Prepare(insertMigrationSQL)
	if err != nil {
		panic(fmt.Sprintf("Could not create prepared statement for migration: %v", err))
	}

	for _, m := range migrations {
		var schemas []string
		if m.MigrationType == types.MigrationTypeTenantMigration || m.MigrationType == types.MigrationTypeTenantScript {
			for _, t := range tenants {
				schemas = append(schemas, t.Name)
			}
		} else {
			schemas = []string{filepath.Base(m.SourceDir)}
		}

		for _, s := range schemas {
			common.LogInfo(bc.ctx, "Applying migration type: %d, schema: %s, file: %s ", m.MigrationType, s, m.File)

			if mode != types.ModeTypeSync {
				contents := strings.Replace(m.Contents, schemaPlaceHolder, s, -1)
				if _, err = tx.Exec(contents); err != nil {
					panic(fmt.Sprintf("SQL migration %v failed with error: %v", m.File, err.Error()))
				}
			}

			if _, err = tx.Stmt(insert).Exec(m.Name, m.SourceDir, m.File, m.MigrationType, s, m.Contents, m.CheckSum, versionID); err != nil {
				panic(fmt.Sprintf("Failed to add migration entry: %v", err.Error()))
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
