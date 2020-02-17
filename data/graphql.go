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
  enum MigrationType {
    SingleMigration
    TenantMigration
    SingleScript
    TenantScript
  }
  type Tenant {
    name: String!
  }
	type Query {
    sourceMigrations(name: String, sourceDir: String, file: String, migrationType: MigrationType): [SourceMigration!]!
    sourceMigration(file: String!): SourceMigration!
    tenants(): [Tenant!]!
	}
`

type sourceMigrationsFilters struct {
	Name          *string
	SourceDir     *string
	File          *string
	MigrationType *types.MigrationType
}

// RootResolver is resolver for all the migrator data
type RootResolver struct {
	Coordinator coordinator.Coordinator
}

func (r *RootResolver) Tenants() ([]types.Tenant, error) {
	tenants := r.Coordinator.GetTenants()
	return tenants, nil
}

func (r *RootResolver) SourceMigrations(args sourceMigrationsFilters) ([]types.Migration, error) {
	allSourceMigrations := r.Coordinator.GetSourceMigrations()
	filteredMigrations := r.filterMigrations(allSourceMigrations, args)
	return filteredMigrations, nil
}

func (r *RootResolver) SourceMigration(args struct {
	File string
}) (types.Migration, error) {
	filter := sourceMigrationsFilters{File: &args.File}
	allSourceMigrations := r.Coordinator.GetSourceMigrations()
	filteredMigrations := r.filterMigrations(allSourceMigrations, filter)
	if len(filteredMigrations) == 0 {
		return types.Migration{}, fmt.Errorf("Source migration %q not found", args.File)
	}
	return filteredMigrations[0], nil
}

func (r *RootResolver) filterMigrations(migrations []types.Migration, args sourceMigrationsFilters) []types.Migration {
	filtered := []types.Migration{}
	ps := reflect.ValueOf(&args)
	s := ps.Elem()
	for _, m := range migrations {
		match := true
		for i := 0; i < s.Type().NumField(); i++ {
			// if filter is nil it means match
			if s.Field(i).IsNil() {
				continue
			}
			// args field names match migration names
			pm := reflect.ValueOf(m).FieldByName(s.Type().Field(i).Name)

			// custom Scalar types are mapped to struct wrappers (so that pointers work correctly during unmarshalling)
			if s.Field(i).Elem().Type().Kind() == reflect.Struct {
				match = match && (pm.Interface() == s.Field(i).Elem().Field(0).Interface())
			} else {
				// other Scalar types are compared directly
				match = match && (pm.Interface() == s.Field(i).Elem().Interface())
			}
			// if already non match don't bother further looping
			if !match {
				continue
			}
		}
		if match {
			filtered = append(filtered, m)
		}
	}
	return filtered
}
