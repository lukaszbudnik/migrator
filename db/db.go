package db

import (
	"context"
	"database/sql"
	"fmt"
	"path/filepath"
	"sort"
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
	GetVersionByID(ID int32) (*types.Version, error)
	GetDBMigrationByID(ID int32) (*types.DBMigration, error)
	GetAppliedMigrations() []types.DBMigration
	CreateVersion(string, types.Action, []types.Migration, bool) (*types.Summary, *types.Version)
	CreateTenant(string, string, types.Action, []types.Migration, bool) (*types.Summary, *types.Version)
	HealthCheck() error
	Dispose()
}

// baseConnector struct is a base struct for implementing DB specific dialects
type baseConnector struct {
	ctx         context.Context
	config      *config.Config
	dialect     dialect
	db          *sql.DB
	initialised bool
}

// Factory is a factory method for creating Loader instance
type Factory func(context.Context, *config.Config) Connector

// New constructs Connector instance based on the passed Config
func New(ctx context.Context, config *config.Config) Connector {
	if config.Driver == "mongodb" {
		return newMongoDBConnector(ctx, config)
	}
	dialect := newDialect(config)
	connector := &baseConnector{ctx, config, dialect, nil, false}
	return connector
}

const (
	migratorSchema           = "migrator"
	migratorTenantsTable     = "migrator_tenants"
	migratorMigrationsTable  = "migrator_migrations"
	migratorVersionsTable    = "migrator_versions"
	defaultSchemaPlaceHolder = "{schema}"
)

// init initialises migrator by making sure proper schema/table are created
func (bc *baseConnector) init() error {
	if bc.initialised {
		return nil
	}
	if bc.db == nil {
		db, err := sql.Open(bc.config.Driver, bc.config.DataSource)
		if err != nil {
			return fmt.Errorf("failed to open connection to DB: %v", err.Error())
		}
		bc.db = db

		if err := bc.db.Ping(); err != nil {
			return fmt.Errorf("failed to connect to database: %v", err)
		}
	}

	tx, err := bc.db.Begin()
	if err != nil {
		return fmt.Errorf("could not start DB transaction: %v", err)
	}

	// make sure migrator schema exists
	createSchema := bc.dialect.GetCreateSchemaSQL(migratorSchema)
	if _, err := bc.db.Exec(createSchema); err != nil {
		return fmt.Errorf("could not create migrator schema: %v", err)
	}

	// make sure migrations table exists
	createMigrationsTable := bc.dialect.GetCreateMigrationsTableSQL()
	if _, err := bc.db.Exec(createMigrationsTable); err != nil {
		return fmt.Errorf("could not create migrations table: %v", err)
	}

	// make sure versions table exists
	createVersionsTableSQLs := bc.dialect.GetCreateVersionsTableSQL()
	for _, createVersionsTableSQL := range createVersionsTableSQLs {
		if _, err := bc.db.Exec(createVersionsTableSQL); err != nil {
			return fmt.Errorf("could not create versions table: %v", err)
		}
	}

	// if using default migrator tenants table make sure it exists
	if bc.config.TenantSelectSQL == "" {
		createTenantsTable := bc.dialect.GetCreateTenantsTableSQL()
		if _, err := bc.db.Exec(createTenantsTable); err != nil {
			return fmt.Errorf("could not create default tenants table: %v", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("could not commit transaction: %v", err)
	}

	bc.initialised = true

	return nil
}

func (bc *baseConnector) initOrPanic() {
	if err := bc.init(); err != nil {
		panic(fmt.Sprintf("Error initialising migrator: %v", err))
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
	bc.initOrPanic()

	tenantSelectSQL := bc.getTenantSelectSQL()

	tenants := []types.Tenant{}

	rows, err := bc.db.Query(tenantSelectSQL)
	if err != nil {
		panic(fmt.Sprintf("Could not query tenants: %v", err))
	}
	defer rows.Close()

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
	bc.initOrPanic()

	versionsSelectSQL := bc.dialect.GetVersionsSelectSQL()

	rows, err := bc.db.Query(versionsSelectSQL)
	if err != nil {
		panic(fmt.Sprintf("Could not query versions: %v", err))
	}
	defer rows.Close()

	return bc.readVersions(rows)
}

func (bc *baseConnector) GetVersionsByFile(file string) []types.Version {
	bc.initOrPanic()

	versionsSelectSQL := bc.dialect.GetVersionsByFileSQL()

	rows, err := bc.db.Query(versionsSelectSQL, file)
	if err != nil {
		panic(fmt.Sprintf("Could not query versions: %v", err))
	}
	defer rows.Close()

	return bc.readVersions(rows)
}

func (bc *baseConnector) GetVersionByID(ID int32) (*types.Version, error) {
	bc.initOrPanic()

	versionsSelectSQL := bc.dialect.GetVersionByIDSQL()

	rows, err := bc.db.Query(versionsSelectSQL, ID)
	if err != nil {
		panic(fmt.Sprintf("Could not query versions: %v", err))
	}
	defer rows.Close()

	// readVersions is generic and returns a slice of Version objects
	// we are querying by ID and are interested in only the first one
	versions := bc.readVersions(rows)

	if len(versions) == 0 {
		return nil, fmt.Errorf("version not found ID: %v", ID)
	}

	return &versions[0], nil
}

func (bc *baseConnector) getVersionByIDInTx(tx *sql.Tx, ID int32) *types.Version {
	versionsSelectSQL := bc.dialect.GetVersionByIDSQL()

	rows, err := tx.Query(versionsSelectSQL, ID)
	if err != nil {
		panic(fmt.Sprintf("Could not query versions: %v", err))
	}
	defer rows.Close()

	// readVersions is generic and returns a slice of Version objects
	// we are querying by ID and are interested in only the first one
	versions := bc.readVersions(rows)

	// when running in transaction version must be found
	if len(versions) == 0 {
		panic(fmt.Sprintf("Version not found ID: %v", ID))
	}

	return &versions[0]
}

func (bc *baseConnector) readVersions(rows *sql.Rows) []types.Version {
	versions := []types.Version{}
	versionsMap := map[int64]*types.Version{}

	for rows.Next() {
		var (
			vid           int64
			vname         string
			vcreated      time.Time
			mid           int64
			name          string
			sourceDir     string
			filename      string
			migrationType types.MigrationType
			schema        string
			created       time.Time
			contents      string
			checksum      string
		)

		if err := rows.Scan(&vid, &vname, &vcreated, &mid, &name, &sourceDir, &filename, &migrationType, &schema, &created, &contents, &checksum); err != nil {
			panic(fmt.Sprintf("Could not read versions: %v", err))
		}
		if versionsMap[vid] == nil {
			version := types.Version{ID: int32(vid), Name: vname, Created: graphql.Time{Time: vcreated}}
			versionsMap[vid] = &version
		}

		version := versionsMap[vid]
		migration := types.Migration{Name: name, SourceDir: sourceDir, File: filename, MigrationType: migrationType, Contents: contents, CheckSum: checksum}
		version.DBMigrations = append(version.DBMigrations, types.DBMigration{Migration: migration, ID: int32(mid), Schema: schema, Created: graphql.Time{Time: created}})
	}

	// map to versions
	for _, v := range versionsMap {
		versions = append(versions, *v)
	}
	// since we used map above sort by version
	sort.Slice(versions, func(i, j int) bool {
		return versions[i].ID > versions[j].ID
	})
	return versions
}

func (bc *baseConnector) GetDBMigrationByID(ID int32) (*types.DBMigration, error) {
	bc.initOrPanic()

	query := bc.dialect.GetMigrationByIDSQL()

	rows, err := bc.db.Query(query, ID)
	if err != nil {
		panic(fmt.Sprintf("Could not query DB migrations: %v", err.Error()))
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, fmt.Errorf("DB migration not found ID: %v", ID)
	}

	var (
		id            int64
		name          string
		sourceDir     string
		filename      string
		migrationType types.MigrationType
		schema        string
		created       time.Time
		contents      string
		checksum      string
	)
	if err = rows.Scan(&id, &name, &sourceDir, &filename, &migrationType, &schema, &created, &contents, &checksum); err != nil {
		panic(fmt.Sprintf("Could not read DB migration: %v", err.Error()))
	}
	m := types.Migration{Name: name, SourceDir: sourceDir, File: filename, MigrationType: migrationType, Contents: contents, CheckSum: checksum}
	db := types.DBMigration{Migration: m, ID: int32(id), Schema: schema, Created: graphql.Time{Time: created}}

	return &db, nil
}

// GetAppliedMigrations returns a list of all applied DB migrations
func (bc *baseConnector) GetAppliedMigrations() []types.DBMigration {
	bc.initOrPanic()

	query := bc.dialect.GetMigrationSelectSQL()

	dbMigrations := []types.DBMigration{}

	rows, err := bc.db.Query(query)
	if err != nil {
		panic(fmt.Sprintf("Could not query DB migrations: %v", err.Error()))
	}
	defer rows.Close()

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
		dbMigrations = append(dbMigrations, types.DBMigration{Migration: mdef, Schema: schema, Created: graphql.Time{Time: created}})
	}
	return dbMigrations
}

// CreateVersion creates new DB version and applies passed migrations
func (bc *baseConnector) CreateVersion(versionName string, action types.Action, migrations []types.Migration, dryRun bool) (*types.Summary, *types.Version) {
	if len(migrations) == 0 {
		return &types.Summary{
			StartedAt: graphql.Time{Time: time.Now()},
			Duration:  0,
		}, nil
	}

	tenants := bc.GetTenants()

	tx, err := bc.db.Begin()
	if err != nil {
		panic(fmt.Sprintf("Could not start transaction: %v", err.Error()))
	}

	defer func() {
		r := recover()
		if r == nil {
			if dryRun {
				common.LogInfo(bc.ctx, "Running in dry-run mode, calling rollback")
				tx.Rollback()
			} else {
				common.LogInfo(bc.ctx, "Running %v, committing transaction", action)
				if err := tx.Commit(); err != nil {
					panic(fmt.Sprintf("Could not commit transaction: %v", err.Error()))
				}
			}
		} else {
			common.LogInfo(bc.ctx, "Recovered in CreateVersion. Transaction rollback.")
			tx.Rollback()
			panic(r)
		}
	}()

	results := bc.applyMigrationsInTx(tx, versionName, action, tenants, migrations)
	version := bc.getVersionByIDInTx(tx, results.VersionID)

	return results, version
}

// CreateTenant creates new tenant and applies passed tenant migrations
func (bc *baseConnector) CreateTenant(tenant string, versionName string, action types.Action, migrations []types.Migration, dryRun bool) (*types.Summary, *types.Version) {
	bc.initOrPanic()

	tenantInsertSQL := bc.getTenantInsertSQL()

	tx, err := bc.db.Begin()
	if err != nil {
		panic(fmt.Sprintf("Could not start transaction: %v", err.Error()))
	}

	defer func() {
		r := recover()
		if r == nil {
			if dryRun {
				common.LogInfo(bc.ctx, "Running in dry-run mode, calling rollback")
				tx.Rollback()
			} else {
				common.LogInfo(bc.ctx, "Running %v action, committing transaction", action)
				if err := tx.Commit(); err != nil {
					panic(fmt.Sprintf("Could not commit transaction: %v", err.Error()))
				}
			}
		} else {
			common.LogInfo(bc.ctx, "Recovered in CreateTenant. Transaction rollback.")
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
	results := bc.applyMigrationsInTx(tx, versionName, action, []types.Tenant{tenantStruct}, migrations)

	version := bc.getVersionByIDInTx(tx, results.VersionID)

	return results, version
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

func (bc *baseConnector) applyMigrationsInTx(tx *sql.Tx, versionName string, action types.Action, tenants []types.Tenant, migrations []types.Migration) *types.Summary {

	results := &types.Summary{
		StartedAt: graphql.Time{Time: time.Now()},
		Tenants:   int32(len(tenants)),
	}

	defer func() {
		results.Duration = time.Since(results.StartedAt.Time).Seconds()
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
		result, _ := stmt.Exec(versionName)
		versionID, _ = result.LastInsertId()
	} else {
		stmt.QueryRow(versionName).Scan(&versionID)
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
			common.LogDebug(bc.ctx, "Applying migration type: %d, schema: %s, file: %s ", m.MigrationType, s, m.File)

			if action == types.ActionApply {
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
			results.TenantMigrationsTotal += int32(len(schemas))
		}
		if m.MigrationType == types.MigrationTypeTenantScript {
			results.TenantScripts++
			results.TenantScriptsTotal += int32(len(schemas))
		}

	}

	results.VersionID = int32(versionID)

	return results
}

func (bc *baseConnector) HealthCheck() error {
	if err := bc.init(); err != nil {
		return err
	}
	return bc.db.Ping()
}
