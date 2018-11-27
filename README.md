# Migrator [![Build Status](https://travis-ci.org/lukaszbudnik/migrator.svg?branch=master)](https://travis-ci.org/lukaszbudnik/migrator)

DB migration tool written in go.

# Usage

Short and sweet.

```
$ migrator -h
Usage of migrator:
  -action string
       migrator action to apply, valid actions are: ["apply" "config" "diskMigrations" "dbTenants" "dbMigrations"] (default "apply")
  -configFile string
       path to migrator.yaml (default "migrator.yaml")
  -mode string
       migrator mode to run: "tool" or "server" (default "tool")
```

Migrator requires a simple `migrator.yaml` file:

```
baseDir: test/migrations
driver: postgres
dataSource: "user=postgres dbname=migrator_test host=192.168.99.100 port=55432 sslmode=disable"
# override only if you have a specific way of determining tenants, default is:
tenantsSql: "select name from public.migrator_tenants"
# override only if you have a specific schema placeholder, default is:
schemaPlaceHolder: {schema}
singleSchemas:
  - public
  - ref
  - config
tenantSchemas:
  - tenants
# port is used only when migrator is run in server mode
# optional element and defaults to 8080
port: 8181
# optional Slack Incoming Web Hook - every apply action posts a message to Slack
slackWebHook: https://hooks.slack.com/services/TTT/BBB/XXX
```

Migrator will scan all directories under `baseDir` directory. Migrations listed under `singleSchemas` directories will be applied once. Migrations listed under `tenantSchemas` directories will be applied for all tenants fetched using `tenantsSql`.

SQL migrations in both `singleSchemas` and `tenantsSchemas` can use `{schema}` placeholder which is automatically replaced by migrator to the current schema. For example:

```
create table if not exists {schema}.modules ( k int, v text );
insert into {schema}.modules values ( 123, '123' );
```

# Server mode

When migrator is run with `-mode server` it starts a go HTTP server and exposes simple REST API which you can use to invoke migrator actions remotely.

All actions which you can invoke from command line can be invoked via REST API:

```
curl http://localhost:8080/config
curl http://localhost:8080/diskMigrations
curl http://localhost:8080/dbTenants
curl http://localhost:8080/dbMigrations
curl -X POST http://localhost:8080/apply
```

Port is configurable in `migrator.yaml` and defaults to 8080. Should you need HTTPS capabilities I encourage you to use nginx/apache/haproxy for SSL/TLS offloading.

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

# Customisation

If you have an existing way of storing information about your tenants you can configure migrator to use it.
In the config file you need to provide 2 parameters:

* `tenantSelectSql` - a select statement which returns names of the tenants
* `tenantInsertSql` - an insert statement which creates a new tenant entry, this is called as a prepared statement and is called with the name of the tenant as a parameter; should your table require additional columns you need to provide default values for them

Here is an example:

```
tenantSelectSql: select name from global.customers
tenantInsertSql: insert into global.customers (name, active, date_added) values (?, true, NOW())
```

# Performance

In my company we use proprietary Ruby framework for DB migrations. I used that framework as a benchmark.

I used performance test generator shipped with migrator (see `test/performance`).

Results are following:

| # Tenants 	| # Existing Migrations 	| # Migrations to apply 	| Migrator time 	| Ruby time 	| Migrator times faster 	|
|-----------	|-----------------------	|-----------------------	|---------------	|-----------	|-----------------------	|
|        10 	|                     0 	|                 10001 	|          177s 	|      670s 	|                  3.78 	|
|        10 	|                 10001 	|                    20 	|            2s 	|      455s 	|                 227.5 	|

Note: The Ruby framework has a pretty undesired functionality of making a DB call to check if given migration was already applied. Migrator fetches all applied migrations at once and compares them in memory. This is the primary reason why migrator is so much better. The other thing to consider is the fact that migrator is written in golang which is known to be much faster than Ruby.

# Installation and supported Go versions

To install migrator use:

`go get github.com/lukaszbudnik/migrator`

Migrator supports the following Go versions: 1.8, 1.9, 1.10, 1.11 (all built on Travis).

# Code Style

If you would like to send me a pull request please always add unit/integration tests. Code should be formatted & checked using the following commands:

```
$ gofmt -s -w .
$ golint ./...
$ go tool vet -v .
```

# License

Copyright 2016 Łukasz Budnik

Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except in compliance with the License. You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the specific language governing permissions and limitations under the License.
