# migrator [![Build Status](https://travis-ci.org/lukaszbudnik/migrator.svg?branch=master)](https://travis-ci.org/lukaszbudnik/migrator) [![Go Report Card](https://goreportcard.com/badge/github.com/lukaszbudnik/migrator)](https://goreportcard.com/report/github.com/lukaszbudnik/migrator) [![codecov](https://codecov.io/gh/lukaszbudnik/migrator/branch/master/graph/badge.svg)](https://codecov.io/gh/lukaszbudnik/migrator)

Super fast and lightweight DB migration & evolution tool written in go.

migrator manages all the DB changes for you and completely eliminates manual and error-prone administrative tasks. migrator not only supports single schemas, but also comes with a multi-tenant support.

migrator can run as a HTTP REST service. Further, there is a ready-to-go migrator docker image.

# Usage

migrator exposes a simple REST API which you can use to invoke different actions:

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
curl http://localhost:8080/migrations
curl -X POST http://localhost:8080/migrations
curl -X POST -H "Content-Type: application/json" -d '{"name": "new_tenant"}' http://localhost:8080/tenants
```

Port is configurable in `migrator.yaml` and defaults to 8080. Should you need HTTPS capabilities I encourage you to use nginx/apache/haproxy for TLS offloading.

There is an official docker image available on docker hub. migrator docker image is ultra lightweight and has a size of 15MB. Ideal for micro-services deployments!

To find out more about migrator docker container see [DOCKER.md](DOCKER.md) for more details.

# Configuration

migrator requires a simple `migrator.yaml` file:

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
# optional webhook configuration section
# URL and template are required if at least one of them is empty noop notifier is used
webHookURL: https://hooks.slack.com/services/TTT/BBB/XXX
# the {text} placeholder is replaced by migrator with information about executed migrations or added new tenant
webHookTemplate: "{\"text\": \"{text}\",\"icon_emoji\": \":white_check_mark:\"}"
# should you need more control over HTTP headers use below
webHookHeaders:
  - "Authorization: Basic QWxhZGRpbjpPcGVuU2VzYW1l"
  - "Content-Type: application/json"
  - "X-CustomHeader: value1,value2"
```

migrator supports env variables substitution in config file. All patterns matching `${NAME}` will look for env variable `NAME`. Below are some common use cases:

```yaml
dataSource: "user=${DB_USER} password=${DB_PASSWORD} dbname=${DB_NAME} host=${DB_HOST} port=${DB_PORT}"
webHookHeaders:
  - "X-Security-Token: ${SECURITY_TOKEN}"
```

# migrator under the hood

migrator scans all directories under `baseDir` directory. Migrations listed under `singleSchemas` directories will be applied once. Migrations listed under `tenantSchemas` directories will be applied for all tenants fetched using `tenantSelectSQL`.

SQL migrations in both `singleSchemas` and `tenantsSchemas` can use `{schema}` placeholder which will be automatically replaced by migrator with a current schema. For example:

```sql
create schema if not exists {schema};
create table if not exists {schema}.modules ( k int, v text );
insert into {schema}.modules values ( 123, '123' );
```

When using migrator please remember about these:

* migrator creates `migrator` schema (where `migrator_migrations` and `migrator_tenants` tables reside) automatically
* when adding a new tenant migrator creates a new schema automatically
* single schemas are not created automatically, for this you must add initial migration with `create schema` SQL statement (see example above)

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

# 2 minutes walkthrough

You can run your first migrations with migrator in literally couple minutes. There are some test migrations which are placed in `test/migrations` directory as well as some docker scripts for setting up test databases.

Let's start.

1. Clone migrator project locally

Cloning project will fetch test migrations and test docker scripts.

For running migrator on docker `git clone` is enough:

```
git fetch origin git@github.com:lukaszbudnik/migrator.git
cd migrator
```

For building migrator from source code `go get` is required:

```
go get github.com/lukaszbudnik/migrator
cd $GOPATH/src/github.com/lukaszbudnik/migrator
```

Setup test DB container. Let's use postgres (see `ultimate-coverage.sh` for all supported containers).

```
./test/docker/create-and-setup-container.sh postgres
```

The script apart of starting test DB container also generates a ready-to-use test config file. We will use it later.

2. a. Build and run migrator from source

```
./setup.sh
go build
./migrator -configFile test/migrator.yaml
```

2. b. Run migrator from docker

Provide a link to `migrator-postgres`. We also need to update `migrator.yaml`:

```
sed -i "s/host=[^ ]* port=[^ ]*/host=migrator-postgres port=5432/g" test/migrator.yaml
sed -i "s/baseDir: .*/baseDir: \/data\/migrations/g" test/migrator.yaml
docker pull lukasz/migrator
docker run -p 8080:8080 -v $PWD/test:/data -e MIGRATOR_YAML=/data/migrator.yaml -d --link migrator-postgres lukasz/migrator
```

3. Play around with migrator

```
curl http://localhost:8080/
curl http://localhost:8080/diskMigrations
curl http://localhost:8080/tenants
curl http://localhost:8080/migrations
curl -X POST http://localhost:8080/migrations
curl -X POST -H "Content-Type: application/json" -d '{"name": "new_tenant"}' http://localhost:8080/tenants
```

Break sha256 checksum of first migration and try to apply migrations or add new tenant.

```
echo " " >> test/migrations/config/201602160001.sql
curl -X POST http://localhost:8080/migrations
curl -X POST -H "Content-Type: application/json" -d '{"name": "new_tenant2"}' http://localhost:8080/tenants
```

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

| # Tenants 	| # Existing Migrations 	| # Migrations to apply 	| migrator 	| Ruby       	| Flyway   	|
|-----------	|-----------------------	|-----------------------	|----------	|-----------	|----------	|
|        10 	|                     0 	|                 10001 	|     154s 	|      670s 	|    2360s 	|
|        10 	|                 10001 	|                    20 	|       2s 	|      455s 	|     340s 	|

migrator is the undisputed winner.

The Ruby framework has an undesired functionality of making a DB call each time to check if given migration was already applied. migrator fetches all applied migrations at once and compares them in memory. This is the primary reason why migrator is so much better in the second test.

flyway results are... very surprising. I was so shocked that I had to re-run flyway as well as all other tests. Yes, flyway is 15 times slower than migrator in the first test. In the second test flyway was faster than Ruby. Still a couple orders of magnitude slower than migrator.

The other thing to consider is the fact that migrator is written in go which is known to be much faster than Ruby and Java.

# Installation and supported Go versions

To install migrator use:

```
go get github.com/lukaszbudnik/migrator
cd migrator
./setup.sh
```

migrator supports the following Go versions: 1.8, 1.9, 1.10, and 1.11 (all built on Travis).

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
