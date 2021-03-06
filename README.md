# migrator [![Build Status](https://travis-ci.org/lukaszbudnik/migrator.svg?branch=master)](https://travis-ci.org/lukaszbudnik/migrator) [![Go Report Card](https://goreportcard.com/badge/github.com/lukaszbudnik/migrator)](https://goreportcard.com/report/github.com/lukaszbudnik/migrator) [![codecov](https://codecov.io/gh/lukaszbudnik/migrator/branch/master/graph/badge.svg)](https://codecov.io/gh/lukaszbudnik/migrator)

Super fast and lightweight DB migration tool written in go. migrator consumes 6MB of memory and outperforms other DB migration/evolution frameworks by a few orders of magnitude.

migrator manages and versions all the DB changes for you and completely eliminates manual and error-prone administrative tasks. migrator versions can be used for auditing and compliance purposes. migrator not only supports single schemas, but also comes with a multi-schema support (ideal for multi-schema multi-tenant SaaS products).

migrator runs as a HTTP REST service and can be easily integrated into your continuous integration and continuous delivery pipeline.

The official docker image is available on docker hub at [lukasz/migrator](https://hub.docker.com/r/lukasz/migrator). It is ultra lightweight and has a size of 15MB. Ideal for micro-services deployments!

# Table of contents

* [API](#api)
  * [/v2 - GraphQL API](#v2---graphql-api)
    * [GET /v2/config](#get-v2config)
    * [GET /v2/schema](#get-v2schema)
    * [POST /v2/service](#post-v2service)
  * [/v1 - REST API](#v1---rest-api)
  * [Request tracing](#request-tracing)
* [Quick Start Guide](#quick-start-guide)
  * [1. Get the migrator project](#1-get-the-migrator-project)
  * [2. Start test DB containers](#2-start-test-db-containers)
  * [3. Run migrator from official docker image](#3-run-migrator-from-official-docker-image)
  * [4. Build and run migrator](#4-build-and-run-migrator)
  * [5. Play around with migrator](#5-play-around-with-migrator)
* [Configuration](#configuration)
  * [migrator.yaml](#migratoryaml)
  * [Env variables substitution](#env-variables-substitution)
  * [Source migrations](#source-migrations)
    * [Local storage](#local-storage)
    * [AWS S3](#aws-s3)
    * [Azure Blob](#azure-blob)
  * [Supported databases](#supported-databases)
* [Customisation and legacy frameworks support](#customisation-and-legacy-frameworks-support)
  * [Custom tenants support](#custom-tenants-support)
  * [Custom schema placeholder](#custom-schema-placeholder)
  * [Synchonising legacy migrations to migrator](#synchonising-legacy-migrations-to-migrator)
  * [Final comments](#final-comments)
* [Tutorials](#tutorials)
  * [Deploying migrator to AWS ECS](#deploying-migrator-to-aws-ecs)
  * [Deploying migrator to AWS EKS](#deploying-migrator-to-aws-eks)
  * [Deploying migrator to Azure AKS](#deploying-migrator-to-azure-aks)
  * [Securing migrator with OAuth2](#securing-migrator-with-oauth2)
  * [Securing migrator with OIDC](#securing-migrator-with-oidc)
* [Performance](#performance)
* [Change log](#change-log)
* [Contributing, code style, running unit & integration tests](#contributing-code-style-running-unit--integration-tests)
* [License](#license)

# API

migrator exposes a REST and GraphQL APIs described below.

To return build information together with a list of supported API versions execute:

```
curl -v http://localhost:8080/
```

Sample HTTP response:

```
< HTTP/1.1 200 OK
< Content-Type: application/json; charset=utf-8
< Date: Mon, 02 Mar 2020 19:48:45 GMT
< Content-Length: 150
<
{"release":"dev-v2020.1.0","commitSha":"c871b176f6e428e186dfe5114a9c86d52a4350f2","commitDate":"2020-03-01T20:58:32+01:00","apiVersions":["v1","v2"]}
```

## /v2 - GraphQL API

API v2 was introduced in migrator v2020.1.0. API v2 is a GraphQL API.

API v2 introduced a formal concept of a DB version. Every migrator action creates a new DB version. Version logically groups all applied DB migrations for auditing and compliance purposes. You can browse versions together with executed DB migrations using the GraphQL API.

## GET /v2/config

Returns migrator's config as `application/x-yaml`.

Sample request:

```
curl -v http://localhost:8080/v2/config
```

Sample HTTP response:

```
< HTTP/1.1 200 OK
< Content-Type: application/x-yaml; charset=utf-8
< Date: Mon, 02 Mar 2020 20:03:13 GMT
< Content-Length: 244
<
baseLocation: test/migrations
driver: sqlserver
dataSource: sqlserver://SA:YourStrongPassw0rd@127.0.0.1:32774/?database=migratortest&connection+timeout=1&dial+timeout=1
singleMigrations:
- ref
- config
tenantMigrations:
- tenants
pathPrefix: /
```

## GET /v2/schema

Returns migrator's GraphQL schema as `plain/text`.

Although migrator supports GraphQL introspection it is much more convenient to get the schema in the plain text.

Sample request:

```
curl -v http://localhost:8080/v2/schema
```

The API v2 GraphQL schema and its description is as follows:

```graphql
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
```

## POST /v2/service

This is a GraphQL endpoint which handles both query and mutation requests.

There are code generators available which can generate client code based on GraphQL schema. This would be the preferred way of consuming migrator's GraphQL endpoint.

In [Quick Start Guide](#quick-start-guide) there are a few curl examples to get you started.

## /v1 - REST API

**As of migrator v2020.1.0 API v1 is deprecated and will sunset in v2021.1.0.**

API v1 is available in migrator v4.x and v2020.x.

The documentation is available in a separate document [API v1](APIv1.md).

## Request tracing

migrator uses request tracing via `X-Request-ID` header. This header can be used with all requests for tracing and/or auditing purposes. If this header is absent migrator will generate one for you.

# Quick Start Guide

You can apply your first migrations with migrator in literally a couple of minutes. There are some test migrations which are located in `test` directory as well as docker-compose for setting up test databases.

The quick start guide shows you how to either use the official docker image or build migrator locally.

## 1. Get the migrator project

Get the source code the usual go way:

```
go get -d -v github.com/lukaszbudnik/migrator
cd $GOPATH/src/github.com/lukaszbudnik/migrator
```

migrator aims to support 3 latest Go versions (built automatically on Travis).

## 2. Start test DB containers

Start and setup test DB containers:

```
docker-compose -f ./test/dokcer-compose.yaml up
```

docker-compose will start and configure 5 different database containers (3 MySQL flavours, PostgreSQL, and MSSQL). Every database container has a ready-to-use migrator config in `test` directory.

## 3. Run migrator from official docker image

The official migrator docker image is available on docker hub [lukasz/migrator](https://hub.docker.com/r/lukasz/migrator).

In order to run the docker container remember to:

1. mount a volume with migrations, for example: `/data`
2. specify location of migrator configuration file, for convenience it is usually located under `/data` directory; it defaults to `/data/migrator.yaml` and can be overridden by setting environment variable `MIGRATOR_YAML`

The docker-compose which starts test DB containers also starts latest migrator with sample configuration and test migrations. See `docker-compose.yaml` for details.

## 4. Build and run migrator

migrator uses go modules to manage dependencies. When building & running migrator from source code simply execute:

```
go build
./migrator -configFile test/migrator-postgresql.yaml
```

> Note: There are 3 git variables injected into the production build. When migrator is built like above it prints empty branch/tag, commit sha, and commit date. This is OK for local development. If you want to inject proper values take a look at `Dockerfile` for details.

## 5. Play around with migrator

If you started migrator in point 3 - migrator listens on port 8181 and connects to mysql server.
If you started migrator in point 4 - migrator listens on port 8080 and connects to postgresql server.

Set the port accordingly:

```
MIGRATOR_PORT=8181
```

Create new version, return version id and name together with operation summary:

```
# versionName parameter is required and can be:
# 1. your version number
# 2. if you do multiple deploys to dev envs perhaps it could be a version number concatenated with current date time
# 3. or if you do CI/CD the commit sha (recommended)
COMMIT_SHA="acfd70fd1f4c7413e558c03ed850012627c9caa9"
# new lines are used for readability but have to be removed from the actual request
cat <<EOF | tr -d "\n" > create_version.txt
{
  "query": "
  mutation CreateVersion(\$input: VersionInput!) {
    createVersion(input: \$input) {
      version {
        id,
        name,
      }
      summary {
        startedAt
        tenants
        migrationsGrandTotal
        scriptsGrandTotal
      }
    }
  }",
  "operationName": "CreateVersion",
  "variables": {
    "input": {
      "versionName": "$COMMIT_SHA"
    }
  }
}
EOF
# and now execute the above query
curl -d @create_version.txt http://localhost:$MIGRATOR_PORT/v2/service
```

Create new tenant, run in dry-run mode, run `Sync` action (instead of default `Apply`), return version id and name, DB migrations, and operation summary:

```
# versionName parameter is required and can be:
# 1. your version number
# 2. if you do multiple deploys to dev envs perhaps it could be a version number concatenated with current date time
# 3. or if you do CI/CD the commit sha (recommended)
COMMIT_SHA="acfd70fd1f4c7413e558c03ed850012627c9caa9"
# tenantName parameter is also required (should not come as a surprise since we want to create new tenant)
TENANT_NAME="new_customer_of_yours"
# new lines are used for readability but have to be removed from the actual request
cat <<EOF | tr -d "\n" > create_tenant.txt
{
  "query": "
  mutation CreateTenant(\$input: TenantInput!) {
    createTenant(input: \$input) {
      version {
        id,
        name,
        dbMigrations {
          id,
          file,
          schema
        }
      }
      summary {
        startedAt
        tenants
        migrationsGrandTotal
        scriptsGrandTotal
      }
    }
  }",
  "operationName": "CreateTenant",
  "variables": {
    "input": {
      "dryRun": true,
      "action": "Sync",
      "versionName": "$COMMIT_SHA - $TENANT_NAME",
      "tenantName": "$TENANT_NAME"
    }
  }
}
EOF
# and now execute the above query
curl -d @create_tenant.txt http://localhost:$MIGRATOR_PORT/v2/service
```

Migrator supports multiple operations in a single GraphQL query. Let's fetch source single migrations, source tenant migrations, and tenants in a single GraphQL query:

```
# new lines are used for readability but have to be removed from the actual request
cat <<EOF | tr -d "\n" > query.txt
{
  "query": "
  query Data(\$singleMigrationsFilters: SourceMigrationFilters, \$tenantMigrationsFilters: SourceMigrationFilters) {
    singleTenantSourceMigrations: sourceMigrations(filters: \$singleMigrationsFilters) {
      file
      migrationType
    }
    multiTenantSourceMigrations: sourceMigrations(filters: \$tenantMigrationsFilters) {
      file
      migrationType
      checkSum
    }
    tenants {
      name
    }
  }",
  "operationName": "Data",
  "variables": {
    "singleMigrationsFilters": {
      "migrationType": "SingleMigration"
    },
    "tenantMigrationsFilters": {
      "migrationType": "TenantMigration"
    }
  }
}
EOF
# and now execute the above query
curl -d @query.txt http://localhost:$MIGRATOR_PORT/v2/service
```

For more GraphQL query and mutation examples see `data/graphql_test.go`.

# Configuration

Let's see how to configure migrator.

## migrator.yaml

migrator configuration file is a simple YAML file. Take a look at a sample `migrator.yaml` configuration file which contains the description, correct syntax, and sample values for all available properties.

```yaml
# required, location where all migrations are stored, see singleSchemas and tenantSchemas below
baseLocation: test/migrations
# required, SQL go driver implementation used, see section "Supported databases"
driver: postgres
# required, dataSource format is specific to SQL go driver implementation used, see section "Supported databases"
dataSource: "user=postgres dbname=migrator_test host=192.168.99.100 port=55432 sslmode=disable"
# optional, override only if you have a specific way of determining tenants, default is:
tenantSelectSQL: "select name from migrator.migrator_tenants"
# optional, override only if you have a specific way of creating tenants, default is:
tenantInsertSQL: "insert into migrator.migrator_tenants (name) values ($1)"
# optional, override only if you have a specific schema placeholder, default is:
schemaPlaceHolder: {schema}
# required, directories of single schema SQL migrations, these are subdirectories of baseLocation
singleMigrations:
  - public
  - ref
  - config
# optional, directories of tenant schemas SQL migrations, these are subdirectories of baseLocation
tenantMigrations:
  - tenants
# optional, directories of single SQL scripts which are applied always, these are subdirectories of baseLocation
singleScripts:
  - config-scripts
# optional, directories of tenant SQL script which are applied always for all tenants, these are subdirectories of baseLocation
tenantScripts:
  - tenants-scripts
# optional, default is:
port: 8080
# path prefix is optional and defaults to '/'
# path prefix is used for application HTTP request routing by Application Load Balancers/Application Gateways
# for example when deploying to AWS ECS and using AWS ALB the path prefix could set as below
# then all HTTP requests should be prefixed with that path, for example: /migrator/v1/config, /migrator/v1/migrations/source, etc.
pathPrefix: /migrator
# the webhook configuration section is optional
# the default Content-Type header is application/json but can be overridden via webHookHeaders below
webHookURL: https://your.server.com/services/TTT/BBB/XXX
# if the webhook expects a payload in a specific format there is an option to provide a payload template
# see webhook template for more information
webHookTemplate: '{"text": "New version: ${summary.versionId} started at: ${summary.startedAt} and took ${summary.duration}. Full results are: ${summary}"}'
# should you need more control over HTTP headers use below
webHookHeaders:
  - "Authorization: Basic QWxhZGRpbjpPcGVuU2VzYW1l"
  - "Content-Type: application/json"
  - "X-CustomHeader: value1,value2"
```

## Env variables substitution

migrator supports env variables substitution in config file. All patterns matching `${NAME}` will look for env variable `NAME`. Below are some common use cases:

```yaml
dataSource: "user=${DB_USER} password=${DB_PASSWORD} dbname=${DB_NAME} host=${DB_HOST} port=${DB_PORT}"
webHookHeaders:
  - "X-Security-Token: ${SECURITY_TOKEN}"
```

## WebHook template

By default when a webhook is configured migrator will post a JSON representation of `Summary` struct to its endpoint.

If your webhook expects a payload in a specific format (say Slack or MS Teams incoming webhooks) there is an option to configure a `webHookTemplate` property in migrator's configuration file. The template can have the following placeholders:

* `${summary}` - will be replaced by a JSON representation of `Summary` struct, all double quotes will be escaped so that the template remains a valid JSON document
* `${summary.field}` - will be replaced by a given field of `Summary` struct

Placeholders can be mixed:

```yaml
webHookTemplate: '{"text": "New version created: ${summary.versionId} started at: ${summary.startedAt} and took ${summary.duration}. Migrations/scripts total: ${summary.migrationsGrandTotal}/${summary.scriptsGrandTotal}. Full results are: ${summary}"}'
```

## Source migrations

Migrations can be read from local disk, AWS S3, Azure Blob Containers. I'm open to contributions to add more cloud storage options.

### Local storage

If `baseLocation` property is a path (either relative or absolute) local storage implementation is used:

```
# relative path
baseLocation: test/migrations
# absolute path
baseLocation: /project/migrations
```

### AWS S3

If `baseLocation` starts with `s3://` prefix, AWS S3 implementation is used. In such case the `baseLocation` property is treated as a bucket name:

```
# S3 bucket
baseLocation: s3://your-bucket-migrator
```

migrator uses official AWS SDK for Go and uses a well known [default credential provider chain](https://docs.aws.amazon.com/sdk-for-go/v1/developer-guide/configuring-sdk.html). Please setup your env variables accordingly.

### Azure Blob Containers

If `baseLocation` matches `^https://.*\.blob\.core\.windows\.net/.*` regex, Azure Blob implementation is used. In such case the `baseLocation` property is treated as a container URL:

```
# Azure Blob container URL
baseLocation: https://storageaccountname.blob.core.windows.net/mycontainer
```

migrator uses official Azure Blob SDK for Go. Unfortunately as of the time of writing Azure Blob implementation the SDK only supported authentication using Storage Accounts and not for example much more flexible Active Directory (which is supported by the rest of the Azure Go SDK). Issue to watch: [Authorization via Azure AD / RBAC](https://github.com/Azure/azure-storage-blob-go/issues/160). I plan to revisit the authorization once Azure team updates their Azure Blob SDK.

## Supported databases

Currently migrator supports the following databases and their flavours. Please review the Go driver implementation for information about supported features and how `dataSource` configuration property should look like:

* PostgreSQL 9.3+ - schema-based multi-tenant database, with transactions spanning DDL statements, driver used: https://github.com/lib/pq
  * PostgreSQL
  * Amazon RDS PostgreSQL - PostgreSQL-compatible relational database built for the cloud
  * Amazon Aurora PostgreSQL - PostgreSQL-compatible relational database built for the cloud
  * Google CloudSQL PostgreSQL - PostgreSQL-compatible relational database built for the cloud
* MySQL 5.6+ - database-based multi-tenant database, transactions do not span DDL statements, driver used: https://github.com/go-sql-driver/mysql
  * MySQL
  * MariaDB - enhanced near linearly scalable multi-master MySQL
  * Percona - an enhanced drop-in replacement for MySQL
  * Amazon RDS MySQL - MySQL-compatible relational database built for the cloud
  * Amazon Aurora MySQL - MySQL-compatible relational database built for the cloud
  * Google CloudSQL MySQL - MySQL-compatible relational database built for the cloud
* Microsoft SQL Server - a relational database management system developed by Microsoft, driver used: https://github.com/denisenkom/go-mssqldb
  * Microsoft SQL Server 2017
  * Microsoft SQL Server 2019

# Customisation and legacy frameworks support

migrator can be used with an already existing legacy DB migration framework.

## Custom tenants support

If you have an existing way of storing information about your tenants you can configure migrator to use it.
In the config file you need to provide 2 configuration properties:

* `tenantSelectSQL` - a select statement which returns names of the tenants
* `tenantInsertSQL` - an insert statement which creates a new tenant entry, the insert statement should be a valid prepared statement for the SQL driver/database you use, it must accept the name of the new tenant as a parameter; finally should your table require additional columns you need to provide default values for them

Here is an example:

```yaml
tenantSelectSQL: select name from global.customers
tenantInsertSQL: insert into global.customers (name, active, date_added) values (?, true, NOW())
```

## Custom schema placeholder

SQL migrations and scripts can use `{schema}` placeholder which will be automatically replaced by migrator with a current schema. For example:

```sql
create schema if not exists {schema};
create table if not exists {schema}.modules ( k int, v text );
insert into {schema}.modules values ( 123, '123' );
```

If you have an existing DB migrations legacy framework which uses different schema placeholder you can override the default one.
In the config file you need to provide `schemaPlaceHolder` configuration property:

For example:

```yaml
schemaPlaceHolder: :tenant
```

## Synchonising legacy migrations to migrator

Before switching from a legacy tool you need to synchronise source migrations to migrator. migrator has no knowledge of migrations applied by other tools and as such will attempt to apply all found source migrations.

Synchronising will load all source migrations and mark them as applied. This can be done by `CreateVersion` operation with action set to `Sync`.

Once the initial synchronisation is done you can use migrator for all the consecutive DB migrations.

## Final comments

When using migrator please remember that:

* migrator creates `migrator` schema together with `migrator_versions` and `migrator_migrations` tables automatically
* if you're not using [Custom tenants support](#custom-tenants-support) migrator creates `migrator_tenants` table automatically
* when adding a new tenant migrator creates a new DB schema and applies all tenant migrations and scripts
* single schemas are not created automatically, you must add initial migration with `create schema {schema}` SQL statement (see sample migrations in `test` folder)

# Tutorials

In this section I provide links to more in-depth migrator tutorials.

## Deploying migrator to AWS ECS

The goal of this tutorial is to deploy migrator to AWS ECS, load migrations from AWS S3 and apply them to AWS RDS DB while storing env variables securely in AWS Secrets Manager. The list of all AWS services used is: IAM, ECS, ECR, Secrets Manager, RDS, and S3.

You can find it in [tutorials/aws-ecs](tutorials/aws-ecs).

## Deploying migrator to AWS EKS

The goal of this tutorial is to deploy migrator to AWS EKS, load migrations from AWS S3 and apply them to AWS RDS DB. The list of AWS services used is: IAM, EKS, ECR, RDS, and S3.

You can find it in [tutorials/aws-eks](tutorials/aws-eks).

## Deploying migrator to Azure AKS

The goal of this tutorial is to publish migrator image to Azure ACR private container repository, deploy migrator to Azure AKS, load migrations from Azure Blob Container and apply them to Azure Database for PostgreSQL. The list of Azure services used is: AKS, ACR, Blob Storage, and Azure Database for PostgreSQL.

You can find it in [tutorials/azure-aks](tutorials/azure-aks).

## Securing migrator with OAuth2

The goal of this tutorial is to secure migrator with OAuth2. It shows how to deploy oauth2-proxy in front of migrator which will off-load and transparently handle authorization for migrator end-users.

You can find it in [tutorials/oauth2-proxy](tutorials/oauth2-proxy).

## Securing migrator with OIDC

The goal of this tutorial is to secure migrator with OAuth2 and OIDC. It shows how to deploy oauth2-proxy and haproxy in front of migrator which will off-load and transparently handle both authorization (oauth2-proxy) and authentication (haproxy with custom lua script) for migrator end-users.

You can find it in [tutorials/oauth2-proxy-oidc-haproxy](tutorials/oauth2-proxy-oidc-haproxy).

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

# Change log

Please navigate to [migrator/releases](https://github.com/lukaszbudnik/migrator/releases) for a complete list of versions, features, and change log.

# Contributing, code style, running unit & integration tests

Contributions are most welcomed.

If you would like to help me and implement a new feature, enhance existing one, or spotted and fixed bug please send me a pull request.

Code should be formatted, checked, and tested using the following commands:

```
./fmt-lint-vet.sh
./coverage.sh
```

The `db/db_integration_test.go` uses go subtests and runs all tests agains 5 different database containers (3 MySQL flavours, PostgreSQL, and MSSQL).

# License

Copyright 2016-2020 Łukasz Budnik

Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except in compliance with the License. You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the specific language governing permissions and limitations under the License.
