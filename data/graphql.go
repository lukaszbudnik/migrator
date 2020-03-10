package data

import (
	"github.com/lukaszbudnik/migrator/coordinator"
	"github.com/lukaszbudnik/migrator/types"
)

// SchemaDefinition contains GraphQL migrator schema
const SchemaDefinition = `
schema {
  query: Query
  mutation: Mutation
}
enum MigrationType {
  SingleMigration
  TenantMigration
  SingleScript
  TenantScript
}
enum Action {
  // Apply is the default action, migrator reads all source migrations and applies them
  Apply
  // Sync is an action where migrator reads all source migrations and marks them as applied in DB
  // typical use cases are:
  // importing source migrations from a legacy tool or synchronising tenant migrations when tenant was created using external tool
  Sync
}
scalar Time
interface Migration {
  name: String!
  migrationType: MigrationType!
  sourceDir: String!
  file: String!
  contents: String!
  checkSum: String!
}
type SourceMigration implements Migration {
  name: String!
  migrationType: MigrationType!
  sourceDir: String!
  file: String!
  contents: String!
  checkSum: String!
}
type DBMigration implements Migration {
  id: Int!
  name: String!
  migrationType: MigrationType!
  sourceDir: String!
  file: String!
  contents: String!
  checkSum: String!
  schema: String!
  created: Time!
}
type Tenant {
  name: String!
}
type Version {
  id: Int!
  name: String!
  created: Time!
  dbMigrations: [DBMigration!]!
}
input SourceMigrationFilters {
  name: String
  sourceDir: String
  file: String
  migrationType: MigrationType
}
input VersionInput {
  versionName: String!
  action: Action = Apply
  dryRun: Boolean = false
}
input TenantInput {
  tenantName: String!
  versionName: String!
  action: Action = Apply
  dryRun: Boolean = false
}
type Summary {
  // date time operation started
  startedAt: Time!
  // how long the operation took in nanoseconds
  duration: Int!
  // number of tenants in the system
  tenants: Int!
  // number of applied single schema migrations
  singleMigrations: Int!
  // number of applied multi-tenant schema migrations
  tenantMigrations: Int!
  // number of all applied multi-tenant schema migrations (equals to tenants * tenantMigrations)
  tenantMigrationsTotal: Int!
  // sum of singleMigrations and tenantMigrationsTotal
  migrationsGrandTotal: Int!
  // number of applied single schema scripts
  singleScripts: Int!
  // number of applied multi-tenant schema scripts
  tenantScripts: Int!
  // number of all applied multi-tenant schema migrations (equals to tenants * tenantScripts)
  tenantScriptsTotal: Int!
  // sum of singleScripts and tenantScriptsTotal
  scriptsGrandTotal: Int!
}
type CreateResults {
  summary: Summary!
  version: Version
}
type Query {
  // returns array of SourceMigration objects
  // note that if input query includes contents field this operation can produce large amounts of data - see sourceMigration(file: String!)
  // all parameters are optional and can be used to filter source migrations
  sourceMigrations(filters: SourceMigrationFilters): [SourceMigration!]!
  // returns a single SourceMigration
  // this operation can be used to fetch a complete SourceMigration including its contents field
  // file is the unique identifier for a source migration which you can get from sourceMigrations() operation
  sourceMigration(file: String!): SourceMigration
  // returns array of Version objects
  // note that if input query includes DBMigration array this operation can produce large amounts of data - see version(id: Int!) or dbMigration(id: Int!)
  // file is optional and can be used to return versions in which given source migration was applied
  versions(file: String): [Version!]!
  // returns a single Version
  // note that if input query includes contents field this operation can produce large amounts of data - see dbMigration(id: Int!)
  // id is the unique identifier of a version which you can get from versions() operation
  version(id: Int!): Version
  // returns a single DBMigration
  // this operation can be used to fetch a complete SourceMigration including its contents field
  // id is the unique identifier of a version which you can get from versions(file: String) or version(id: Int!) operations
  dbMigration(id: Int!): DBMigration
  // returns array of Tenant objects
  tenants(): [Tenant!]!
}
type Mutation {
  // creates new DB version by applying all eligible DB migrations & scripts
  createVersion(input: VersionInput!): CreateResults!
  // creates new tenant by applying only tenant-specific DB migrations & scripts, also creates new DB version
  createTenant(input: TenantInput!): CreateResults!
}
`

// RootResolver is resolver for all the migrator data
type RootResolver struct {
	Coordinator coordinator.Coordinator
}

// Tenants resolves all tenants
func (r *RootResolver) Tenants() ([]types.Tenant, error) {
	tenants := r.Coordinator.GetTenants()
	return tenants, nil
}

// Versions resoves all versions, optionally can return versions with specific source migration (file is the identifier for source migrations)
func (r *RootResolver) Versions(args struct {
	File *string
}) ([]types.Version, error) {
	if args.File != nil {
		return r.Coordinator.GetVersionsByFile(*args.File), nil
	}
	return r.Coordinator.GetVersions(), nil
}

// Version resolves version by ID
func (r *RootResolver) Version(args struct {
	ID int32
}) (*types.Version, error) {
	return r.Coordinator.GetVersionByID(args.ID)
}

// SourceMigrations resolves source migrations using optional filters
func (r *RootResolver) SourceMigrations(args struct {
	Filters *coordinator.SourceMigrationFilters
}) ([]types.Migration, error) {
	sourceMigrations := r.Coordinator.GetSourceMigrations(args.Filters)
	return sourceMigrations, nil
}

// SourceMigration resolves source migration by its file name
func (r *RootResolver) SourceMigration(args struct {
	File string
}) (*types.Migration, error) {
	return r.Coordinator.GetSourceMigrationByFile(args.File)
}

// DBMigration resolves DB migration by ID
func (r *RootResolver) DBMigration(args struct {
	ID int32
}) (*types.MigrationDB, error) {
	return r.Coordinator.GetDBMigrationByID(args.ID)
}

// CreateVersion creates new DB version
func (r *RootResolver) CreateVersion(args struct {
	Input types.VersionInput
}) (*types.CreateResults, error) {
	results := r.Coordinator.CreateVersion(args.Input.VersionName, args.Input.Action, args.Input.DryRun)
	return results, nil
}

// CreateTenant creates new tenant
func (r *RootResolver) CreateTenant(args struct {
	Input types.TenantInput
}) (*types.CreateResults, error) {
	results := r.Coordinator.CreateTenant(args.Input.VersionName, args.Input.Action, args.Input.DryRun, args.Input.TenantName)
	return results, nil
}
