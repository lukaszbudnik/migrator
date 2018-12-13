# Migrator [![Build Status](https://travis-ci.org/lukaszbudnik/migrator.svg?branch=master)](https://travis-ci.org/lukaszbudnik/migrator) [![Go Report Card](https://goreportcard.com/badge/github.com/lukaszbudnik/migrator)](https://goreportcard.com/report/github.com/lukaszbudnik/migrator) [![codecov](https://codecov.io/gh/lukaszbudnik/migrator/branch/master/graph/badge.svg)](https://codecov.io/gh/lukaszbudnik/migrator)

Super fast and lightweight DB migration & evolution tool written in go.

migrator manages all the DB changes for you and completely eliminates manual and error-prone administrative tasks. migrator not only supports single schemas, but also comes with a multi-tenant support.

migrator can run as a HTTP REST service. Further, there is a ready-to-go migrator docker image.

# Usage

Short and sweet.

```
$ Usage of ./migrator:
  -action string
    	when run in tool mode, action to execute, valid actions are: ["apply" "addTenant" "config" "diskMigrations" "dbTenants" "dbMigrations"] (default "apply")
  -configFile string
    	path to migrator configuration yaml file (default "migrator.yaml")
  -mode string
    	migrator mode to run: ["tool" "server"] (default "tool")
  -tenant string
    	when run in tool mode and action set to "addTenant", specifies new tenant name
```

Migrator requires a simple `migrator.yaml` file:

```yaml
baseDir: test/migrations
driver: postgres
# dataSource format is specific to DB go driver implementation - see below 'Supported databases'
dataSource: "user=postgres dbname=migrator_test host=192.168.99.100 port=55432 sslmode=disable"
# override only if you have a specific way of determining tenants, default is:
tenantSelectSQL: "select name from migrator.migrator_tenants"
# override only if you have a specific way of creating tenants, default is:
tenantInsertSQL: "insert into migrator.migrator_tenants (name) values ($1)"
# override only if you have a specific schema placeholder, default is:
schemaPlaceHolder: {schema}
singleSchemas:
  - public
  - ref
  - config
tenantSchemas:
  - tenants
# port is used only when migrator is run in server mode, defaults to:
port: 8080
# optional Slack Incoming Web Hook - if defined apply migrations action will post a message to Slack
slackWebHook: https://hooks.slack.com/services/TTT/BBB/XXX
```

Migrator will scan all directories under `baseDir` directory. Migrations listed under `singleSchemas` directories will be applied once. Migrations listed under `tenantSchemas` directories will be applied for all tenants fetched using `tenantSelectSQL`.

SQL migrations in both `singleSchemas` and `tenantsSchemas` can use `{schema}` placeholder which will be automatically replaced by migrator with a current schema. For example:

```sql
create schema if not exists {schema};
create table if not exists {schema}.modules ( k int, v text );
insert into {schema}.modules values ( 123, '123' );
```

# DB Schemas

When using migrator please remember about these:

* migrator creates `migrator` schema (where `migrator_migrations` and `migrator_tenants` tables reside) automatically
* when adding a new tenant migrator creates a new schema automatically
* single schemas are not created automatically, for this you must add initial migration with `create schema` SQL statement (see example above)

# Server mode

When migrator is run with `-mode server` it starts a HTTP service and exposes simple REST API which you can use to invoke migrator actions remotely.

All actions which you can invoke from command line can be invoked via REST API:

* GET / - returns migrator config, response is `Content-Type: application/x-yaml`
* GET /diskMigrations - returns disk migrations, response is `Content-Type: application/json`
* GET /tenants - returns tenants, response is `Content-Type: application/json`
* POST /tenants - adds new tenant, name parameter is passed as JSON, returns applied migrations, response is `Content-Type: application/json`
* GET /migrations - returns all applied migrations
* POST /migrations - applies migrations, no parameters required, returns applied migrations, response is `Content-Type: application/json`

Some curl examples to get you started:

```
curl http://localhost:8080/
curl http://localhost:8080/diskMigrations
curl http://localhost:8080/tenants
curl -X POST -H "Content-Type: application/json" -d '{"name": "new_tenant"}' http://localhost:8080/tenants
curl http://localhost:8080/migrations
curl -X POST http://localhost:8080/migrations
```

Port is configurable in `migrator.yaml` and defaults to 8080. Should you need HTTPS capabilities I encourage you to use nginx/apache/haproxy for TLS offloading.

# Supported databases

Currently migrator supports the following databases and their flavours:

* PostgreSQL 9.3+ - schema-based multi-tenant database, with transactions spanning DDL statements, driver used: https://github.com/lib/pq
  * PostgreSQL - original PostgreSQL server
  * Amazon RDS PostgreSQL - PostgreSQL-compatible relational database built for the cloud
  * Amazon Aurora PostgreSQL - PostgreSQL-compatible relational database built for the cloud
  * Google CloudSQL PostgreSQL - PostgreSQL-compatible relational database built for the cloud
* MySQL 5.6+ - database-based multi-tenant database, transactions do not span DDL statements, driver used: https://github.com/go-sql-driver/mysql
  * MySQL - original MySQL server
  * MariaDB - enhanced near linearly scalable multi-master MySQL
  * Percona - an enhanced drop-in replacement for MySQL
  * Amazon RDS MySQL - MySQL-compatible relational database built for the cloud
  * Amazon Aurora MySQL - MySQL-compatible relational database built for the cloud
  * Google CloudSQL MySQL - MySQL-compatible relational database built for the cloud
* Microsoft SQL Server 2017 - a relational database management system developed by Microsoft, driver used: https://github.com/denisenkom/go-mssqldb
  * Microsoft SQL Server - original Microsoft SQL Server

# Do you speak docker?

Yes, there is an official docker image available on docker hub.

migrator docker image is ultra lightweight and has a size of approx. 15MB. Ideal for micro-services deployments!

To find out more about migrator docker container see [DOCKER.md](DOCKER.md) for more details.

# Customisation

If you have an existing way of storing information about your tenants you can configure migrator to use it.
In the config file you need to provide 2 parameters:

* `tenantSelectSQL` - a select statement which returns names of the tenants
* `tenantInsertSQL` - an insert statement which creates a new tenant entry, this is called as a prepared statement and is called with the name of the tenant as a parameter; should your table require additional columns you need to provide default values for them

Here is an example:

```yaml
tenantSelectSQL: select name from global.customers
tenantInsertSQL: insert into global.customers (name, active, date_added) values (?, true, NOW())
```

# Performance

As a benchmarks I used 2 migrations frameworks:

* proprietary Ruby framework - used at my company
* flyway - leading market feature rich DB migration framework: https://flywaydb.org

There is a performance test generator shipped with migrator (`test/performance/generate-test-migrations.sh`). In order to generate flyway-compatible migrations you need to pass `-f` param (see script for details).

Execution times are following:

| # Tenants 	| # Existing Migrations 	| # Migrations to apply 	| Migrator 	| Ruby       	| Flyway   	|
|-----------	|-----------------------	|-----------------------	|----------	|-----------	|----------	|
|        10 	|                     0 	|                 10001 	|     154s 	|      670s 	|    2360s 	|
|        10 	|                 10001 	|                    20 	|       2s 	|      455s 	|     340s 	|

Migrator is the undisputed winner.

The Ruby framework has an undesired functionality of making a DB call each time to check if given migration was already applied. Migrator fetches all applied migrations at once and compares them in memory. This is the primary reason why migrator is so much better in the second test.

flyway results are... very surprising. I was so shocked that I had to re-run flyway as well as all other tests. Yes, flyway is 15 times slower than migrator in the first test. In the second test flyway was faster than Ruby. Still a couple orders of magnitude slower than migrator.

The other thing to consider is the fact that migrator is written in go which is known to be much faster than Ruby and Java.

# Installation and supported Go versions

To install migrator use:

```
go get github.com/lukaszbudnik/migrator
cd migrator
./setup.sh
```

Migrator supports the following Go versions: 1.8, 1.9, 1.10, and 1.11 (all built on Travis).

# Contributing, code style, running unit & integration tests

Contributions are most welcomed.

If you would like to help me and implement a new feature, enhance existing one, or spotted and fixed bug please send me a pull request.

Code should be formatted, checked, and tested using the following commands:

```
./fmt-lint-vet.sh
./ultimate-coverage.sh
```

The `ultimate-coverage.sh` script loops through 5 different containers (3 MySQL flavours, PostgreSQL, and MSSQL) creates db docker container, executes `coverage.sh` script, and finally tears down given db docker container.

# License

Copyright 2016-2018 Łukasz Budnik

Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except in compliance with the License. You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the specific language governing permissions and limitations under the License.
