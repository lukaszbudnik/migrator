# migrator ![Build and Test](https://github.com/lukaszbudnik/migrator/workflows/Build%20and%20Test/badge.svg) ![Docker](https://github.com/lukaszbudnik/migrator/workflows/Docker%20Image%20CI/badge.svg) [![Go Report Card](https://goreportcard.com/badge/github.com/lukaszbudnik/migrator)](https://goreportcard.com/report/github.com/lukaszbudnik/migrator) [![codecov](https://codecov.io/gh/lukaszbudnik/migrator/branch/master/graph/badge.svg)](https://codecov.io/gh/lukaszbudnik/migrator)

Super fast and lightweight DB migration tool written in go. migrator outperforms other DB migration/evolution frameworks by a few orders of magnitude.

migrator manages and versions all the DB changes for you and completely eliminates manual and error-prone administrative tasks. migrator versions can be used for auditing and compliance purposes. migrator not only supports single schemas, but also comes with a multi-schema support (ideal for multi-schema multi-tenant SaaS products).

migrator runs as a HTTP GrapQL service and can be easily integrated into existing continuous integration and continuous delivery pipelines. migrator can also sync existing migrations from legacy frameworks making the technology switch even more straightforward.

migrator supports reading DB migrations from:

- local folder (any Docker/Kubernetes deployments)
- AWS S3
- Azure Blob Containers

migrator support the following multi-tenant databases:

- PostgreSQL 9.3+ (and all its flavours)
- MySQL 5.6+ (and all its flavours)
- Microsoft SQL Server 2017+

The official docker image is available on docker hub at [lukasz/migrator](https://hub.docker.com/r/lukasz/migrator) or on the alternative mirror at [ghcr.io/lukaszbudnik/migrator](https://github.com/lukaszbudnik/migrator/pkgs/container/migrator).
It is ultra lightweight and has a size of 30MB. Ideal for micro-services deployments!

# API

migrator exposes a REST and GraphQL APIs described below.

To return build information together with a list of supported API versions execute:

```bash
curl http://localhost:8080/
```

Sample HTTP response:

```
{"release":"v2020.1.3","commitSha":"b56a2694fcdb523e0c3f3e79b2d7a1b61f28a91f","commitDate":"2020-10-12T13:22:57+02:00","apiVersions":["v1","v2"]}
```

## /v2 - GraphQL API

API v2 is a GraphQL API. API v2 was introduced in migrator v2020.1.0.

API v2 introduced a formal concept of a DB version. Every migrator action creates a new DB version. Version logically groups all applied DB migrations for auditing and compliance purposes. You can browse versions together with executed DB migrations using the GraphQL API.

## GET /v2/config

Returns migrator's config as `application/x-yaml`.

Sample request:

```bash
curl http://localhost:8080/v2/config
```

Sample HTTP response:

```
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

```bash
curl http://localhost:8080/v2/schema
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

API v1 was sunset in v2021.0.0.

API v1 is available in migrator v4.x and v2020.x.

The documentation is available in a separate document [API v1](APIv1.md).

## Request tracing

migrator uses request tracing via `X-Request-ID` header. This header can be used with all requests for tracing and/or auditing purposes. If this header is absent migrator will generate one for you.

# Quick Start Guide

You can apply your first migrations with migrator in literally a few seconds. There is a ready-to-use docker-compose file which sets up migrator and test databases.

## 1. Get the migrator project

Get the source code the usual go way:

```bash
go get -d -v github.com/lukaszbudnik/migrator
cd $GOPATH/src/github.com/lukaszbudnik/migrator
```

migrator aims to support 3 latest Go versions (built automatically on GitHub Actions).

## 2. Start migrator and test DB containers

Start migrator and setup test DB containers using docker-compose:

```bash
docker-compose -f ./test/docker-compose.yaml up
```

docker-compose will start and configure the following services:

1. `migrator` - service using latest official migrator image, listening on port `8181`
2. `migrator-dev` - service built from local branch, listening on port `8282`
3. `postgres` - PostgreSQL service, listening on port `54325`
4. `mysql` - MySQL service, listening on port `3306`
5. `mariadb` - MariaDB (MySQL flavour), listening on port `13306`
6. `percona` - Percona (MySQL flavour), listening on port `23306`
7. `mssql` - MS SQL Server, listening on port `1433`

> Note: Every database container has a ready-to-use migrator config in `test` directory. You can edit `test/docker-compose.yaml` file and switch to a different database. By default `migrator` and `migrator-dev` services use `test/migrator-docker.yaml` which connects to `mysql` service.

## 3. migrator and migrator-dev services

docker-compose will start 2 migrator services. The first one `migrator` will use the latest official migrator docker image from docker hub [lukasz/migrator](https://hub.docker.com/r/lukasz/migrator). The second one `migrator-dev` will be built automatically by docker-compose from your local branch.

In order to run the docker container remember to:

1. mount a volume with migrations, for example: `/data`
2. specify location of migrator configuration file, for convenience it is usually located under `/data` directory; it defaults to `/data/migrator.yaml` and can be overridden by setting environment variable `MIGRATOR_YAML`

The docker-compose will mount volumes with sample configuration and test migrations for you. See `test/docker-compose.yaml` for details.

> Note: For production deployments please see [Tutorials](#tutorials) section. It contains walkthoughs of deployments to AWS ECS, AWS EKS, and Azure AKS.

## 4. Play around with migrator

The docker-compose will start 2 migrator services as listed above. The latest stable migrator version listens on port `8181`. migrator built from the local branch listens on port `8282`.

Set the port accordingly:

```bash
MIGRATOR_PORT=8181
```

Create new version, return version id and name together with operation summary:

```bash
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

```bash
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

```bash
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
  - "X-Custom-Header: value1,value2"
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

- `${summary}` - will be replaced by a JSON representation of `Summary` struct, all double quotes will be escaped so that the template remains a valid JSON document
- `${summary.field}` - will be replaced by a given field of `Summary` struct

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

If `baseLocation` starts with `s3://` prefix, AWS S3 implementation is used. In such case the `baseLocation` property is treated as a bucket name followed by optional prefix:

```
# S3 bucket
baseLocation: s3://your-bucket-migrator
# S3 bucket with optional prefix
baseLocation: s3://your-bucket-migrator/appcodename/prod/artefacts
```

migrator uses official AWS SDK for Go and uses a well known [default credential provider chain](https://docs.aws.amazon.com/sdk-for-go/v1/developer-guide/configuring-sdk.html).

### Azure Blob Containers

If `baseLocation` matches `^https://.*\.blob\.core\.windows\.net/.*` regex, Azure Blob implementation is used. In such case the `baseLocation` property is treated as a container URL. The URL can have optional prefix too:

```
# Azure Blob container URL
baseLocation: https://storageaccountname.blob.core.windows.net/mycontainer
# Azure Blob container URL with optional prefix
baseLocation: https://storageaccountname.blob.core.windows.net/mycontainer/appcodename/prod/artefacts
```

migrator uses official Azure SDK for Go and supports authentication using Storage Account Key (via `AZURE_STORAGE_ACCOUNT` and `AZURE_STORAGE_ACCESS_KEY` env variables) as well as much more flexible (and recommended) Azure Active Directory Managed Identity.

## Supported databases

Currently migrator supports the following databases and their flavours. Please review the Go driver implementation for information about supported features and how `dataSource` configuration property should look like:

- PostgreSQL 9.3+ - schema-based multi-tenant database, with transactions spanning DDL statements, driver used: https://github.com/lib/pq
  - PostgreSQL
  - Amazon RDS PostgreSQL - PostgreSQL-compatible relational database built for the cloud
  - Amazon Aurora PostgreSQL - PostgreSQL-compatible relational database built for the cloud
  - Google CloudSQL PostgreSQL - PostgreSQL-compatible relational database built for the cloud
- MySQL 5.6+ - database-based multi-tenant database, transactions do not span DDL statements, driver used: https://github.com/go-sql-driver/mysql
  - MySQL
  - MariaDB - enhanced near linearly scalable multi-master MySQL
  - Percona - an enhanced drop-in replacement for MySQL
  - Amazon RDS MySQL - MySQL-compatible relational database built for the cloud
  - Amazon Aurora MySQL - MySQL-compatible relational database built for the cloud
  - Google CloudSQL MySQL - MySQL-compatible relational database built for the cloud
- Microsoft SQL Server - a relational database management system developed by Microsoft, driver used: https://github.com/denisenkom/go-mssqldb
  - Microsoft SQL Server 2017
  - Microsoft SQL Server 2019

# Customisation and legacy frameworks support

migrator can be used with an already existing legacy DB migration framework.

## Custom tenants support

If you have an existing way of storing information about your tenants you can configure migrator to use it.
In the config file you need to provide 2 configuration properties:

- `tenantSelectSQL` - a select statement which returns names of the tenants
- `tenantInsertSQL` - an insert statement which creates a new tenant entry, the insert statement should be a valid prepared statement for the SQL driver/database you use, it must accept the name of the new tenant as a parameter; finally should your table require additional columns you need to provide default values for them

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

- migrator creates `migrator` schema together with `migrator_versions` and `migrator_migrations` tables automatically
- if you're not using [Custom tenants support](#custom-tenants-support) migrator creates `migrator_tenants` table automatically
- when adding a new tenant migrator creates a new DB schema and applies all tenant migrations and scripts
- single schemas are not created automatically, you must add initial migration with `create schema {schema}` SQL statement (see sample migrations in `test` folder)

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

- proprietary Ruby framework - used at my company
- flyway - leading market feature rich DB migration framework: https://flywaydb.org

There is a performance test generator shipped with migrator (`test/performance/generate-test-migrations.sh`). In order to generate flyway-compatible migrations you need to pass `-f` param (see script for details).

Execution times are following:

| # Tenants | # Existing Migrations | # Migrations to apply | migrator | Ruby | Flyway |
| --------- | --------------------- | --------------------- | -------- | ---- | ------ |
| 10        | 0                     | 10001                 | 154s     | 670s | 2360s  |
| 10        | 10001                 | 20                    | 2s       | 455s | 340s   |

migrator is the undisputed winner.

The Ruby framework has an undesired functionality of making a DB call each time to check if given migration was already applied. migrator fetches all applied migrations at once and compares them in memory. This is the primary reason why migrator is so much better in the second test.

flyway results are... very surprising. I was so shocked that I had to re-run flyway as well as all other tests. Yes, flyway is 15 times slower than migrator in the first test. In the second test flyway was faster than Ruby. Still a couple orders of magnitude slower than migrator.

The other thing to consider is the fact that migrator is written in go which is known to be much faster than Ruby and Java.

# Change log

Please navigate to [migrator/releases](https://github.com/lukaszbudnik/migrator/releases) for a complete list of versions, features, and change log.

# Contributing

Contributions are most welcomed!

For contributing, code style, running unit & integration tests please see [CONTRIBUTING.md](CONTRIBUTING.md).

# License

Copyright 2016-2021 Łukasz Budnik

Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except in compliance with the License. You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the specific language governing permissions and limitations under the License.
