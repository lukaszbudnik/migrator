package data

import (
	"fmt"
	"reflect"

	"github.com/lukaszbudnik/migrator/coordinator"
	"github.com/lukaszbudnik/migrator/types"
)

// SchemaDefinition contains GraphQL migrator schema
const SchemaDefinition = `
	schema {
		query: Query
	}
  enum MigrationType {
    SingleMigration
    TenantMigration
    SingleScript
    TenantScript
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
	type Query {
    // returns array of SourceMigration objects
    // note that if input query includes contents field this operation can produce large amounts of data - see sourceMigration(file: String!)
    // all parameters are optional and can be used to filter source migrations
    sourceMigrations(name: String, sourceDir: String, file: String, migrationType: MigrationType): [SourceMigration!]!
    // returns a single SourceMigration
    // this operation can be used to fetch a complete SourceMigration including its contents field
    // file is the unique identifier for a source migration which you can get from sourceMigrations() operation
    sourceMigration(file: String!): SourceMigration!
    // returns array of Version objects
    // note that if input query includes DBMigration array this operation can produce large amounts of data - see version(id: Int!) or dbMigration(id: Int!)
    // file is optional and can be used to return versions in which given source migration was applied
    versions(file: String): [Version!]!
    // returns a single Version
    // note that if input query includes contents field this operation can produce large amounts of data - see dbMigration(id: Int!)
    // id is the unique identifier of a version which you can get from versions() operation
    version(id: Int!): Version!
    // returns a single DBMigration
    // this operation can be used to fetch a complete SourceMigration including its contents field
    // id is the unique identifier of a version which you can get from versions(file: String) or version(id: Int!) operations
    //dbMigration(id: Int!): DBMigration!
    // returns array of Tenant objects
    tenants(): [Tenant!]!
	}
`

type sourceMigrationsFilters struct {
	Name          *string
	SourceDir     *string
	File          *string
	MigrationType *types.MigrationType
}

type dbMigrationsFilters struct {
	sourceMigrationsFilters
	Schema *string
}

// RootResolver is resolver for all the migrator data
type RootResolver struct {
	Coordinator coordinator.Coordinator
}

func (r *RootResolver) Tenants() ([]types.Tenant, error) {
	tenants := r.Coordinator.GetTenants()
	return tenants, nil
}

func (r *RootResolver) Versions(args struct {
	File *string
}) ([]types.Version, error) {
	if args.File != nil {
		return r.Coordinator.GetVersionsByFile(*args.File), nil
	}
	return r.Coordinator.GetVersions(), nil
}

func (r *RootResolver) Version(args struct {
	ID int32
}) (types.Version, error) {
	return r.Coordinator.GetVersionByID(args.ID), nil
}

func (r *RootResolver) SourceMigrations(filters sourceMigrationsFilters) ([]types.Migration, error) {
	allSourceMigrations := r.Coordinator.GetSourceMigrations()
	filteredMigrations := r.filterMigrations(allSourceMigrations, filters)
	return filteredMigrations, nil
}

func (r *RootResolver) SourceMigration(args struct {
	File string
}) (types.Migration, error) {
	filters := sourceMigrationsFilters{File: &args.File}
	allSourceMigrations := r.Coordinator.GetSourceMigrations()
	filteredMigrations := r.filterMigrations(allSourceMigrations, filters)
	if len(filteredMigrations) == 0 {
		return types.Migration{}, fmt.Errorf("Source migration %q not found", args.File)
	}
	return filteredMigrations[0], nil
}

// func (r *RootResolver) DBMigrations(filters dbMigrationsFilters) ([]types.MigrationDB, error) {
// 	dbMigrations := r.Coordinator.GetAppliedMigrations()
// 	filteredMigrations := r.filterDBMigrations(dbMigrations, filters)
// 	return filteredMigrations, nil
// }

// func (r *RootResolver) DBMigration(args struct {
// 	File   string
// 	Schema string
// }) (types.MigrationDB, error) {
// 	filters := dbMigrationsFilters{sourceMigrationsFilters: sourceMigrationsFilters{File: &args.File}, Schema: &args.Schema}
// 	dbMigrations := r.Coordinator.GetAppliedMigrations()
// 	filteredMigrations := r.filterDBMigrations(dbMigrations, filters)
// 	if len(filteredMigrations) == 0 {
// 		return types.MigrationDB{}, fmt.Errorf("Source migration %q not found", args.File)
// 	}
// 	return filteredMigrations[0], nil
// }

func (r *RootResolver) filterMigrations(migrations []types.Migration, filters sourceMigrationsFilters) []types.Migration {
	filtered := []types.Migration{}
	for _, m := range migrations {
		match := r.matchMigration(m, filters)
		if match {
			filtered = append(filtered, m)
		}
	}
	return filtered
}

func (r *RootResolver) filterDBMigrations(migrations []types.MigrationDB, filters dbMigrationsFilters) []types.MigrationDB {
	filtered := []types.MigrationDB{}
	for _, m := range migrations {
		match := r.matchDBMigration(m, filters)
		if match {
			filtered = append(filtered, m)
		}
	}
	return filtered
}

func (r *RootResolver) matchMigration(m types.Migration, filters sourceMigrationsFilters) bool {
	ps := reflect.ValueOf(&filters)
	s := ps.Elem()
	match := true
	for i := 0; i < s.Type().NumField(); i++ {
		// if filter is nil it means match
		if s.Field(i).IsNil() {
			continue
		}
		// args field names match migration names
		pm := reflect.ValueOf(m).FieldByName(s.Type().Field(i).Name)
		match = match && (pm.Interface() == s.Field(i).Elem().Interface())
		// if already non match don't bother further looping
		if !match {
			continue
		}
	}
	return match
}

func (r *RootResolver) matchDBMigration(m types.MigrationDB, filters dbMigrationsFilters) bool {
	match := r.matchMigration(m.Migration, filters.sourceMigrationsFilters)
	// if Migration is already different don't bother checking MigrationDB fields
	if !match {
		return match
	}
	ps := reflect.ValueOf(&filters)
	s := ps.Elem()
	for i := 0; i < s.Type().NumField(); i++ {
		// skip embedded struct - already handled by matchMigration()
		if s.Field(i).Kind() == reflect.Struct {
			continue
		}
		// if filter is nil it means match
		if s.Field(i).IsNil() {
			continue
		}
		// args field names match migration names
		pm := reflect.ValueOf(m).FieldByName(s.Type().Field(i).Name)
		match = match && (pm.Interface() == s.Field(i).Elem().Interface())
		// if already non match don't bother further looping
		if !match {
			break
		}
	}
	return match
}
