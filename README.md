# Migrator [![Build Status](https://travis-ci.org/lukaszbudnik/migrator.svg?branch=master)](https://travis-ci.org/lukaszbudnik/migrator)

DB migration tool written in go.

# Usage

Short and sweet.

```
$ migrator -h
Usage of migrator:
  -action string
    	migrator action to apply, valid actions are: ["apply" "listDiskMigrations" "listDBTenants" "listDBMigrations"] (default "apply")
  -configFile string
    	path to migrator.yaml (default "migrator.yaml")
  -verbose
    	set to "true" to print more data to output
```

Migrator requires a simple `migrator.yaml` file:

```
baseDir: test/migrations
driver: postgres
dataSource: "user=postgres dbname=migrator_test host=192.168.99.100 port=55432 sslmode=disable"
# override only if you have own specific way of determining tenants
tenantsSql: "select name from public.migrator_tenants"
singleSchemas:
  - public
  - ref
  - config
tenantSchemas:
  - tenants
```

Migrator will scan all directories under `baseDir` directory. Migrations listed under `singleSchemas` directories will be applied once. Migrations listed under `tenantSchemas` directories will be applied for all tenants fetched using `tenantsSql`.

SQL migrations in both `singleSchemas` and `tenantsSchemas` can use `{schema}` placeholder which is automatically replaced by migrator to the current schema. For example:

```
create table if not exists {schema}.modules ( k int, v text );
insert into {schema}.modules values ( 123, '123' );
```

# Supported databases

Currently migrator supports the following databases:

* PostgreSQL - schema-based multi-tenant database, with transactions spanning DDL statements
* MySQL - database-based multi-tenant database, transactions do not span DDL statements
* MariaDB - enhanced near linearly scalable multi-master MySQL

# Examples

PostgreSQL:

```
$ docker/postgresql-create-and-setup-container.sh
$ ./coverage.sh
$ docker/postgresql-destroy-container.sh
```

MySQL:

```
$ docker/mysql-create-and-setup-container.sh
$ ./coverage.sh
$ docker/mysql-destroy-container.sh
```

MariaDB:

```
$ docker/mariadb-create-and-setup-container.sh
$ ./coverage.sh
$ docker/mariadb-destroy-container.sh
```

Or see `.travis.yml` to see how it's done on Travis.

# Installation and supported Go versions

To install migrator use:

`go get github.com/lukaszbudnik/migrator`

Migrator supports the following Go versions: 1.2, 1.3, 1.4, 1.5, 1.6 (all built on Travis).

# License

Copyright 2016 Łukasz Budnik

Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except in compliance with the License. You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the specific language governing permissions and limitations under the License.
