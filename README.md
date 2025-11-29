# migrator ![Build](https://github.com/lukaszbudnik/migrator/workflows/Build/badge.svg) ![Docker](https://github.com/lukaszbudnik/migrator/workflows/Docker%20Image%20CI/badge.svg) [![Go Report Card](https://goreportcard.com/badge/github.com/lukaszbudnik/migrator)](https://goreportcard.com/report/github.com/lukaszbudnik/migrator) [![codecov](https://codecov.io/gh/lukaszbudnik/migrator/branch/main/graph/badge.svg)](https://codecov.io/gh/lukaszbudnik/migrator)

Super fast and lightweight DB migration tool written in Go. migrator outperforms other market-leading DB migration frameworks by a few orders of magnitude when comparing both execution time and memory consumption (see [PERFORMANCE.md](PERFORMANCE.md)).

## ✨ Key Features

- **🚀 Ultra Performance**: Orders of magnitude faster than other migration tools
- **☁️ Multi-Cloud Storage**: Read migrations from local disk, AWS S3, or Azure Blob Storage
- **🏢 Multi-Tenant Ready**: Built-in support for multi-schema, multi-tenant SaaS applications
- **📡 GraphQL API**: Modern HTTP GraphQL service with comprehensive query capabilities
- **📊 Observability**: Built-in Prometheus metrics and health checks
- **🐳 Container Native**: Ultra-lightweight 30MB Docker image, perfect for microservices
- **🔧 Legacy Migration**: Sync existing migrations from other frameworks seamlessly
- **🎯 CI/CD Ready**: Easy integration into continuous deployment pipelines

## 🗄️ Supported Databases

- **PostgreSQL** 9.6+ (and flavours: Amazon RDS/Aurora, Google CloudSQL)
- **MySQL** 5.7+ (and flavours: MariaDB, TiDB, Percona, Amazon RDS/Aurora, Google CloudSQL)
- **Microsoft SQL Server** 2008+
- **MongoDB** 4.0+

## 📦 Installation

The official docker image is available on:
- Docker Hub: [lukasz/migrator](https://hub.docker.com/r/lukasz/migrator)
- GitHub Container Registry: [ghcr.io/lukaszbudnik/migrator](https://github.com/lukaszbudnik/migrator/pkgs/container/migrator)

```bash
docker pull lukasz/migrator:latest
```

## 🎯 Overview

migrator manages and versions all DB changes for you and completely eliminates manual and error-prone administrative tasks. migrator versions can be used for auditing and compliance purposes. migrator not only supports single schemas, but also comes with multi-schema support out of the box, making it an ideal DB migrations solution for multi-tenant SaaS products.

migrator runs as a HTTP GraphQL service and can be easily integrated into existing continuous integration and continuous delivery pipelines. migrator can also sync existing migrations from legacy frameworks making the technology switch even more straightforward.

migrator supports the following multi-tenant databases:

- PostgreSQL and all its flavours
- MySQL and all its flavours
- Microsoft SQL Server
- MongoDB

migrator supports reading DB migrations from:

- local folder (any Docker/Kubernetes deployments)
- AWS S3
- Azure Blob Containers

The official docker image is available on:

- docker hub at: [lukasz/migrator](https://hub.docker.com/r/lukasz/migrator)
- alternative mirror at: [ghcr.io/lukaszbudnik/migrator](https://github.com/lukaszbudnik/migrator/pkgs/container/migrator)

The image is ultra lightweight and has a size of 30MB. Ideal for micro-services deployments!

## 📋 Table of Contents

- [🚀 Quick Start Guide](#-quick-start-guide)
- [📡 API](#-api)
- [⚙️ Configuration](#-configuration)
- [📁 Source migrations](#-source-migrations)
- [🗄️ Supported databases](#-supported-databases)
- [🔧 Customisation and legacy frameworks support](#-customisation-and-legacy-frameworks-support)
- [📊 Metrics](#-metrics)
- [🏥 Health Checks](#-health-checks)
- [📚 Tutorials](#-tutorials)
- [⚡ Performance](#-performance)

## 🚀 Quick Start Guide

You can apply your first migrations with migrator in literally a few seconds. There is a ready-to-use docker-compose file which sets up migrator and test databases.

### 1. Get the migrator project

Get the source code:

```bash
git clone https://github.com/lukaszbudnik/migrator.git
cd migrator
```

### 2. Start migrator and test DB containers

Start migrator and setup test DB containers using docker-compose:

```bash
docker-compose -f ./test/docker-compose.yaml up
```

docker-compose will start and configure the following services:

1. `migrator` - service using latest official migrator image, listening on port `8181`
2. `migrator-dev` - service built from local branch, listening on port `8282`
3. `postgres` - PostgreSQL service, listening on port `5432`
4. `mysql` - MySQL service, listening on port `3306`
5. `mssql` - MS SQL Server, listening on port `1433`
6. `mongodb` - MongoDB service, listening on port `27017`

### 3. Play around with migrator

Set the port and create your first migration:

```bash
MIGRATOR_PORT=8181
COMMIT_SHA="your-version-here"

# Create new version
curl -s -d @- http://localhost:$MIGRATOR_PORT/v2/service <<EOF | jq
{
  "query": "mutation CreateVersion(\$input: VersionInput!) {
    createVersion(input: \$input) {
      version { id, name }
    }
  }",
  "variables": {
    "input": {
      "versionName": "$COMMIT_SHA"
    }
  }
}
EOF

# Fetch migrator versions
curl -s -d @- http://localhost:$MIGRATOR_PORT/v2/service <<EOF | jq -r ".data.versions"
{
  "query": "
  query Versions {
    versions {
        id,
        name,
        created,
      }
  }",
  "operationName": "Versions"
}
EOF

# Fetch tenants
curl -s -d @- http://localhost:$MIGRATOR_PORT/v2/service <<EOF | jq -r ".data.tenants"
{
  "query": "
  query Tenants {
    tenants {
        name
      }
  }",
  "operationName": "Tenants"
}
EOF

# Create new tenant
TENANT_NAME="newcustomer$RANDOM"
VERSION_NAME="create-tenant-$TENANT_NAME"
curl -s -d @- http://localhost:$MIGRATOR_PORT/v2/service <<EOF | jq -r '.data.createTenant'
{
  "query": "
  mutation CreateTenant(\$input: TenantInput!) {
    createTenant(input: \$input) {
      version {
        id,
        name,
      }
      summary {
        startedAt
        duration
        tenants
        migrationsGrandTotal
        scriptsGrandTotal
      }
    }
  }",
  "operationName": "CreateTenant",
  "variables": {
    "input": {
      "versionName": "$VERSION_NAME",
      "tenantName": "$TENANT_NAME"
    }
  }
}
EOF
```

> **💡 Tip**: For a complete GraphQL schema and production deployment guides, see the [📡 API](#-api) and [📚 Tutorials](#-tutorials) sections below.

## 📡 API

To return build information together with a list of supported API versions execute:

```bash
MIGRATOR_PORT=8181
curl http://localhost:${MIGRATOR_PORT}/
```

Sample HTTP response:

```json
{
  "release": "refs/tags/v2023.0.1",
  "sha": "43a6858b74832c185783969b3afbf7c2547a533a",
  "apiVersions": ["v2"]
}
```

### /v2 - GraphQL API

API v2 is a GraphQL API. API v2 was introduced in migrator v2020.1.0.

API v2 introduced a formal concept of a DB version. Every migrator action creates a new DB version. Version logically groups all applied DB migrations for auditing and compliance purposes. You can browse versions together with executed DB migrations using the GraphQL API.

### GET /v2/config

Returns migrator's config as `application/yaml`.

Sample request:

```bash
curl http://localhost:${MIGRATOR_PORT}/v2/config
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

### GET /v2/schema

Returns migrator's GraphQL schema as `plain/text`.

Although migrator supports GraphQL introspection, it is much more convenient to get the schema in plain text.

Sample request:

```bash
curl http://localhost:${MIGRATOR_PORT}/v2/schema
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
  // how long the operation took in seconds
  duration: Float!
  // number of tenants in the system
  tenants: Int!
  // number of loaded and applied single schema migrations
  singleMigrations: Int!
  // number of loaded multi-tenant schema migrations
  tenantMigrations: Int!
  // number of applied multi-tenant schema migrations (equals to tenants * tenantMigrations)
  tenantMigrationsTotal: Int!
  // sum of singleMigrations and tenantMigrationsTotal
  migrationsGrandTotal: Int!
  // number of loaded and applied single schema scripts
  singleScripts: Int!
  // number of loaded multi-tenant schema scripts
  tenantScripts: Int!
  // number of applied multi-tenant schema migrations (equals to tenants * tenantScripts)
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
  // all parameters are optional and can be used to filter source migrations
  // note that if the input query includes "contents" field this operation can produce large amounts of data
  // if you want to return "contents" field it may be better to get individual source migrations using sourceMigration(file: String!)
  sourceMigrations(filters: SourceMigrationFilters): [SourceMigration!]!
  // returns a single SourceMigration
  // this operation can be used to fetch a complete SourceMigration including "contents" field
  // file is the unique identifier for a source migration file which you can get from sourceMigrations()
  sourceMigration(file: String!): SourceMigration
  // returns array of Version objects
  // file is optional and can be used to return versions in which given source migration file was applied
  // note that if input query includes DBMigration array and "contents" field this operation can produce large amounts of data
  // if you want to return "contents" field it may be better to get individual versions using either
  // version(id: Int!) or even get individual DB migration using dbMigration(id: Int!)
  versions(file: String): [Version!]!
  // returns a single Version
  // id is the unique identifier of a version which you can get from versions()
  // note that if input query includes "contents" field this operation can produce large amounts of data
  // if you want to return "contents" field it may be better to get individual DB migration using dbMigration(id: Int!)
  version(id: Int!): Version
  // returns a single DBMigration
  // this operation can be used to fetch a complete DBMigration including "contents" field
  // id is the unique identifier of a DB migration which you can get from versions(file: String) or version(id: Int!)
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

### POST /v2/service

This is a GraphQL endpoint which handles both query and mutation requests.

See [Quick Start Guide](#-quick-start-guide) for a few curl examples to get you started.

The preferred way of consuming migrator's GraphQL endpoint is to use GraphQL clients. These clients can be generated from the GraphQL schema in any programming language you use (Java, Python, C#, JavaScript, Go, etc.).

### /v1 - REST API

API v1 was sunset in v2021.0.0.

The documentation is available in a separate document [API v1](APIv1.md).

### Request tracing

migrator uses request tracing via `X-Request-ID` header. This header can be used with all requests for tracing and/or auditing purposes. If this header is absent migrator will generate one for you.

## ⚙️ Configuration

Let's see how to configure migrator.

### migrator.yaml

migrator configuration file is a simple YAML file. Take a look at a sample `migrator.yaml` configuration file which contains the description, correct syntax, and sample values for all available properties.

```yaml
# required, location where all migrations are stored, see singleSchemas and tenantSchemas below
baseLocation: test/migrations
# required, database driver implementation used, see section "Supported databases"
driver: postgres
# required, dataSource format is specific to database driver implementation used, see section "Supported databases"
dataSource: "user=postgres dbname=migrator_test host=192.168.99.100 port=55432 sslmode=disable"
# optional, override only if you have a specific way of determining tenants, default is:
tenantSelectSQL: "select name from migrator.migrator_tenants"
# optional, override only if you have a specific way of creating tenants, default is:
tenantInsertSQL: "insert into migrator.migrator_tenants (name) values ($1)"
# optional, override only if you have a specific schema placeholder, default is:
schemaPlaceHolder: { schema }
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
# optional, directories of tenant SQL scripts which are applied always for all tenants, these are subdirectories of baseLocation
tenantScripts:
  - tenants-scripts
# optional, default is 8080
port: 8080
# path prefix is optional and defaults to '/'
# path prefix is used for application HTTP request routing by Application Load Balancers/Application Gateways
# for example when deploying to AWS ECS and using AWS ALB the path prefix could be set as below
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
# optional, allows to filter logs produced by migrator, valid values are: DEBUG, INFO, ERROR, PANIC
# defaults to INFO
logLevel: INFO
```

### Env variables substitution

migrator supports env variables substitution in config file. All patterns matching `${NAME}` will look for env variable `NAME`. Below are some common use cases:

```yaml
dataSource: "user=${DB_USER} password=${DB_PASSWORD} dbname=${DB_NAME} host=${DB_HOST} port=${DB_PORT}"
webHookHeaders:
  - "X-Security-Token: ${SECURITY_TOKEN}"
```

### WebHook template

By default when a webhook is configured migrator will post a JSON representation of `Summary` struct to its endpoint.

If your webhook expects a payload in a specific format (say Slack or MS Teams incoming webhooks) there is an option to configure a `webHookTemplate` property in migrator's configuration file. The template can have the following placeholders:

- `${summary}` - will be replaced by a JSON representation of `Summary` struct, all double quotes will be escaped so that the template remains a valid JSON document
- `${summary.field}` - will be replaced by a given field of `Summary` struct

Placeholders can be mixed:

```yaml
webHookTemplate: '{"text": "New version created: ${summary.versionId} started at: ${summary.startedAt} and took ${summary.duration}. Migrations/scripts total: ${summary.migrationsGrandTotal}/${summary.scriptsGrandTotal}. Full results are: ${summary}"}'
```

## 📁 Source migrations

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

## 🗄️ Supported databases

Currently migrator supports the following databases including their flavours (like Percona, MariaDB for MySQL, etc.). Please review the Go driver implementation for information about all supported features and how `dataSource` configuration property should look like.

### PostgreSQL 9.6+

Schema-based multi-tenant database, with transactions spanning DDL statements, driver used: https://github.com/lib/pq.

The following versions and flavours are supported by the driver:

- PostgreSQL
- Amazon RDS PostgreSQL - PostgreSQL-compatible relational database built for the cloud
- Amazon Aurora PostgreSQL - PostgreSQL-compatible relational database built for the cloud
- Google CloudSQL PostgreSQL - PostgreSQL-compatible relational database built for the cloud

### MySQL 5.7+

Database-based multi-tenant database, transactions do not span DDL statements, driver used: https://github.com/go-sql-driver/mysql.

The following versions and flavours are supported by the driver:

- MySQL
- MariaDB - enhanced near linearly scalable multi-master MySQL
- TiDB - an open-source, cloud-native, distributed SQL database designed for high availability, horizontal and vertical scalability, strong consistency, and high performance.
- Percona - an enhanced drop-in replacement for MySQL
- Amazon RDS MySQL - MySQL-compatible relational database built for the cloud
- Amazon Aurora MySQL - MySQL-compatible relational database built for the cloud
- Google CloudSQL MySQL - MySQL-compatible relational database built for the cloud

### Microsoft SQL Server 2008+

A relational database management system developed by Microsoft, driver used: https://github.com/microsoft/go-mssqldb.

The Go driver supports all Microsoft SQL Server versions starting with 2008.

### MongoDB 4.0+

Document-oriented NoSQL database, driver used: https://github.com/mongodb/mongo-go-driver.

MongoDB uses database-based multi-tenancy (similar to MySQL). Migrations are JavaScript files executed via MongoDB's eval command. The migrator metadata (versions, migrations, tenants) is stored in a dedicated `migrator` database.

Sample MongoDB configuration:

```yaml
baseLocation: migrations
driver: mongodb
dataSource: "mongodb://localhost:27017"
singleMigrations:
  - admin
tenantMigrations:
  - tenants
```

Sample migration file (`001_create_users.js`):

```javascript
db.users.createIndex({ email: 1 }, { unique: true });
db.users.insertOne({ name: "admin", role: "admin" });
```

## 🔧 Customisation and legacy frameworks support

migrator can be used with an already existing legacy DB migration framework.

### Custom tenants support

If you have an existing way of storing information about your tenants you can configure migrator to use it.
In the config file you need to provide 2 configuration properties:

- `tenantSelectSQL` - a select statement which returns names of the tenants
- `tenantInsertSQL` - an insert statement which creates a new tenant entry, the insert statement should be a valid prepared statement for the SQL driver/database you use, it must accept the name of the new tenant as a parameter; finally should your table require additional columns you need to provide default values for them

Here is an example:

```yaml
tenantSelectSQL: select name from global.customers
tenantInsertSQL: insert into global.customers (name, active, date_added) values (?, true, NOW())
```

### Custom schema placeholder

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

### Synchronising legacy migrations to migrator

Before switching from a legacy tool you need to synchronise source migrations to migrator. migrator has no knowledge of migrations applied by other tools and as such will attempt to apply all found source migrations.

Synchronising will load all source migrations and mark them as applied. This can be done by `CreateVersion` operation with action set to `Sync`.

Once the initial synchronisation is done you can use migrator for all the consecutive DB migrations.

### Final comments

When using migrator please remember that:

- migrator creates `migrator` schema together with `migrator_versions` and `migrator_migrations` tables automatically
- if you're not using [Custom tenants support](#custom-tenants-support) migrator creates `migrator_tenants` table automatically
- when adding a new tenant migrator creates a new DB schema and applies all tenant migrations and scripts
- single schemas are not created automatically, you must add an initial migration with `create schema {schema}` SQL statement (see sample migrations in `test` folder)

## 📊 Metrics

migrator exposes Prometheus metrics at `/metrics` endpoint. Apart from migrator-specific metrics, it exposes a lot of OS process and Go metrics.

The following metrics are available:

- `go_gc_*` - Go garbage collection
- `go_memstats_*` - Go memory
- `process_*` - OS process
- `migrator_gin_request_*` - Gin request metrics
- `migrator_gin_response_*` - Gin response metrics
- `migrator_gin_tenants_created` - migrator tenants created
- `migrator_gin_versions_created` - migrator versions created
- `migrator_gin_migrations_applied{type="single_migrations"}` - migrator single migrations applied
- `migrator_gin_migrations_applied{type="single_scripts"}` - migrator single scripts applied
- `migrator_gin_migrations_applied{type="tenant_migrations_total"}` - migrator total tenant migrations applied (for all tenants)
- `migrator_gin_migrations_applied{type="tenant_scripts_total"}` - migrator total tenant scripts applied (for all tenants)

## 🏥 Health Checks

Health checks are available at `/health` endpoint. migrator implements [Eclipse MicroProfile Health 3.0 RC4](https://download.eclipse.org/microprofile/microprofile-health-3.0-RC4/microprofile-health-spec.html) spec.

A successful response returns HTTP 200 OK code:

```json
{
  "status": "UP",
  "checks": [
    {
      "name": "DB",
      "status": "UP"
    },
    {
      "name": "Loader",
      "status": "UP"
    }
  ]
}
```

In case one of the checks has DOWN status then the overall status is DOWN. Failed check has `data` field which provides more information on why its status is DOWN. Health check will also return HTTP 503 Service Unavailable code:

```json
{
  "status": "DOWN",
  "checks": [
    {
      "name": "DB",
      "status": "DOWN",
      "data": {
        "details": "failed to connect to database: dial tcp 127.0.0.1:5432: connect: connection refused"
      }
    },
    {
      "name": "Loader",
      "status": "DOWN",
      "data": {
        "details": "open /nosuchdir/migrations: no such file or directory"
      }
    }
  ]
}
```

## 📚 Tutorials

In this section I provide links to more in-depth migrator tutorials.

### Deploying migrator to AWS ECS

The goal of this tutorial is to deploy migrator to AWS ECS, load migrations from AWS S3 and apply them to AWS RDS DB while storing env variables securely in AWS Secrets Manager. The list of all AWS services used is: IAM, ECS, ECR, Secrets Manager, RDS, and S3.

You can find it in [tutorials/aws-ecs](tutorials/aws-ecs).

### Deploying migrator to AWS EKS

The goal of this tutorial is to deploy migrator to AWS EKS, load migrations from AWS S3 and apply them to AWS RDS DB. The list of AWS services used is: IAM, EKS, ECR, RDS, and S3.

You can find it in [tutorials/aws-eks](tutorials/aws-eks).

### Deploying migrator to Azure AKS

The goal of this tutorial is to publish migrator image to Azure ACR private container repository, deploy migrator to Azure AKS, load migrations from Azure Blob Container and apply them to Azure Database for PostgreSQL. The list of Azure services used is: AKS, ACR, Blob Storage, and Azure Database for PostgreSQL.

You can find it in [tutorials/azure-aks](tutorials/azure-aks).

### Securing migrator with OAuth2

The goal of this tutorial is to secure migrator with OAuth2. It shows how to deploy oauth2-proxy in front of migrator which will off-load and transparently handle authorization for migrator end-users.

You can find it in [tutorials/oauth2-proxy](tutorials/oauth2-proxy).

### Securing migrator with OIDC

The goal of this tutorial is to secure migrator with OAuth2 and OIDC. It shows how to deploy oauth2-proxy and haproxy in front of migrator which will off-load and transparently handle both authorization (oauth2-proxy) and authentication (haproxy with custom lua script) for migrator end-users.

You can find it in [tutorials/oauth2-proxy-oidc-haproxy](tutorials/oauth2-proxy-oidc-haproxy).

## ⚡ Performance

Performance benchmarks were moved to a dedicated [PERFORMANCE.md](PERFORMANCE.md) document.

## Change log

Please navigate to [migrator/releases](https://github.com/lukaszbudnik/migrator/releases) for a complete list of versions, features, and change log.

## Contributing

Contributions are most welcomed!

For contributing, code style, running unit & integration tests please see [CONTRIBUTING.md](CONTRIBUTING.md).

## License

Apache 2.0 License - see [LICENSE](LICENSE).
