package data

import (
	"fmt"
	"reflect"

	"github.com/lukaszbudnik/migrator/loader"
	"github.com/lukaszbudnik/migrator/types"
)

const schemaString = `
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
	}
`

type sourceMigrationsFilters struct {
	Name          *string
	SourceDir     *string
	File          *string
	MigrationType *types.MigrationType
}

type RootResolver struct {
	loader loader.Loader
}

func (r *RootResolver) SourceMigrations(args sourceMigrationsFilters) ([]types.Migration, error) {
	allSourceMigrations := r.loader.GetSourceMigrations()
	filteredMigrations := r.filterMigrations(allSourceMigrations, args)
	return filteredMigrations, nil
}

func (r *RootResolver) SourceMigration(args struct {
	File string
}) (types.Migration, error) {
	filter := sourceMigrationsFilters{File: &args.File}
	allSourceMigrations := r.loader.GetSourceMigrations()
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
