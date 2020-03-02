# migrator [![Build Status](https://travis-ci.org/lukaszbudnik/migrator.svg?branch=master)](https://travis-ci.org/lukaszbudnik/migrator) [![Go Report Card](https://goreportcard.com/badge/github.com/lukaszbudnik/migrator)](https://goreportcard.com/report/github.com/lukaszbudnik/migrator) [![codecov](https://codecov.io/gh/lukaszbudnik/migrator/branch/master/graph/badge.svg)](https://codecov.io/gh/lukaszbudnik/migrator)

Super fast and lightweight DB migration tool written in go.

migrator manages and versions all the DB changes for you and completely eliminates manual and error-prone administrative tasks. migrator not only supports single schemas, but also comes with a multi-tenant support.

migrator runs as a HTTP REST service and can be easily integrated into your continuous integration and continuous delivery pipeline.

Further, there is an official docker image available on docker hub. [lukasz/migrator](https://hub.docker.com/r/lukasz/migrator) is ultra lightweight and has a size of 15MB. Ideal for micro-services deployments!

# Table of contents

* [Usage](#usage)
  * [GET /](#get-)
  * [/v2 - GraphQL API](#v2---graphql-api)
    * [GET /v2/config](#get-v2config)
    * [POST /v2/service](#post-v2service)
  * [/v1](#v1)
    * [GET /v1/config](#get-v1config)
    * [GET /v1/migrations/source](#get-v1migrationssource)
    * [GET /v1/migrations/applied](#get-v1migrationsapplied)
    * [POST /v1/migrations](#post-v1migrations)
    * [GET /v1/tenants](#get-v1tenants)
    * [POST /v1/tenants](#post-v1tenants)
  * [Request tracing](#request-tracing)
* [Quick Start Guide](#quick-start-guide)
  * [1. Get the migrator project](#1-get-the-migrator-project)
  * [2. Setup test DB container](#2-setup-test-db-container)
  * [3. Build and run migrator](#3-build-and-run-migrator)
  * [4. Run migrator from official docker image](#4-run-migrator-from-official-docker-image)
  * [5. Play around with migrator](#5-play-around-with-migrator)
* [Tutorials](#tutorials)
  * [Deploying migrator to AWS ECS](#deploying-migrator-to-aws-ecs)
  * [Deploying migrator to AWS EKS](#deploying-migrator-to-aws-eks)
  * [Deploying migrator to Azure AKS](#deploying-migrator-to-azure-aks)
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
* [Performance](#performance)
* [Change log](#change-log)
* [Contributing, code style, running unit & integration tests](#contributing-code-style-running-unit--integration-tests)
* [License](#license)

# Usage

migrator exposes a simple REST API described below.

## GET /

Migrator returns build information together with a list of supported API versions.

Sample request:

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

API v2 also introduced a formal concept of DB versions. Every migrator action creates a new DB version. Version logically groups all applied DB migrations for auditing and compliance purposes. You can browse versions together with executed DB migrations using the GraphQL API.

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

Sample HTTP response (truncated):

```
< HTTP/1.1 200 OK
< Content-Type: text/plain; charset=utf-8
< Date: Mon, 02 Mar 2020 20:12:20 GMT
< Transfer-Encoding: chunked
<
schema {
  query: Query
}
enum MigrationType {
  SingleMigration
  TenantMigration
  SingleScript
  TenantScript
}
...
```

## POST /v2/service

This is a GraphQL endpoint which handles both query and mutation requests.

The current GraphQL schema together with description in comments is as follows:

```graphql

```

There are code generators available which can generate client code based on GraphQL schema. This would be the preferred way of consuming migrator's GraphQL endpoint.

Below are a couple of curl examples to get you started.

Create new version:

```
```

Create new tenant:

```
```

Query data:

```
```

For more GraphQL query and mutation examples see `data/graphql_test.go`.

## /v1

**Deprecation**: As of migrator v2020.1.0 API v1 is deprecated and will sunset in v2021.1.0.

API v1 is available in migrator v4.x and v2020.x.

## GET /v1/config

Returns migrator's config as `application/x-yaml`.

Sample request:

```
curl -v http://localhost:8080/v1/config
```

Sample HTTP response:

```
< HTTP/1.1 200 OK
< Content-Type: application/x-yaml; charset=utf-8
< Date: Wed, 01 Jan 2020 17:31:57 GMT
< Content-Length: 277

baseLocation: test/migrations
driver: postgres
dataSource: user=postgres dbname=migrator_test host=127.0.0.1 port=32776 sslmode=disable
  connect_timeout=1
singleMigrations:
- ref
- config
tenantMigrations:
- tenants
singleScripts:
- config-scripts
tenantScripts:
- tenants-scripts
```
## GET /v1/migrations/source

Returns list of all source migrations. Response is a list of JSON representation of `Migration` struct.

Sample request:

```
curl -v http://localhost:8080/v1/migrations/source
```

Sample HTTP response:

```
< HTTP/1.1 200 OK
< Content-Type: application/json; charset=utf-8
< Date: Tue, 31 Dec 2019 11:27:48 GMT
< Transfer-Encoding: chunked

[
  {
    "name": "201602160002.sql",
    "sourceDir": "/Users/lukasz/go/src/github.com/lukaszbudnik/migrator/test/migrations/config",
    "file": "/Users/lukasz/go/src/github.com/lukaszbudnik/migrator/test/migrations/config/201602160002.sql",
    "migrationType": 1,
    "contents": "create table {schema}.config (\n  id integer,\n  k varchar(100),\n  v varchar(100),\n  primary key (id)\n);\n",
    "checkSum": "58db38d8f6c197ab290212470a82fe1f5b1f3cacadbe00ac59cd68a3bfa98baf"
  },
  {
    "name": "201602160002.sql",
    "sourceDir": "/Users/lukasz/go/src/github.com/lukaszbudnik/migrator/test/migrations/tenants",
    "file": "/Users/lukasz/go/src/github.com/lukaszbudnik/migrator/test/migrations/tenants/201602160002.sql",
    "migrationType": 2,
    "contents": "create table {schema}.module (id integer, id_config integer, foreign key (id_config) references config.config(id));\n",
    "checkSum": "56c4c1d8f82f3dedade5116be46267edee01a4889c6359ef03c39dc73ca653a8"
  }
]
```

`Migration` JSON contains the following fields:

* `name` - migration file name
* `sourceDir` - absolute path to source directory
* `file` - absolute path to migration file (concatenation of `sourceDir` and `name`)
* `migrationType` - type of migration, values are:
  * 1 - single migration applied once for a given schema
  * 2 - multi-tenant migration applied once but for all tenants/schemas
  * 3 - single script - special type of migration applied always for a given schema
  * 4 - multi-tenant script - special type of migration applied always for all tenants/schemas
* `contents` - contents of the migration file
* `checkSum` - sha256 checksum of migration file contents

## GET /v1/migrations/applied

Returns list of all applied migrations. Response is a list of JSON representation of `MigrationDB` struct.

Sample request:

```
curl -v http://localhost:8080/v1/migrations/applied
```

Sample HTTP response:

```
< HTTP/1.1 200 OK
< Content-Type: application/json; charset=utf-8
< Date: Wed, 01 Jan 2020 17:32:49 GMT
< Transfer-Encoding: chunked

[
  {
    "name": "201602160001.sql",
    "sourceDir": "/Users/lukasz/go/src/github.com/lukaszbudnik/migrator/test/migrations/config",
    "file": "/Users/lukasz/go/src/github.com/lukaszbudnik/migrator/test/migrations/config/201602160001.sql",
    "migrationType": 1,
    "contents": "create schema config;\n",
    "checkSum": "c1380af7a054ec75778252f539e1e9f914d2c5b1f441ea1df18c2140c6c3380a",
    "schema": "config",
    "appliedAt": "2020-01-01T17:29:13.169306Z"
  },
  {
    "name": "201602160002.sql",
    "sourceDir": "/Users/lukasz/go/src/github.com/lukaszbudnik/migrator/test/migrations/config",
    "file": "/Users/lukasz/go/src/github.com/lukaszbudnik/migrator/test/migrations/config/201602160002.sql",
    "migrationType": 1,
    "contents": "create table {schema}.config (\n  id integer,\n  k varchar(100),\n  v varchar(100),\n  primary key (id)\n);\n",
    "checkSum": "58db38d8f6c197ab290212470a82fe1f5b1f3cacadbe00ac59cd68a3bfa98baf",
    "schema": "config",
    "appliedAt": "2020-01-01T17:29:13.169306Z"
  },
  {
    "name": "201602160002.sql",
    "sourceDir": "/Users/lukasz/go/src/github.com/lukaszbudnik/migrator/test/migrations/tenants",
    "file": "/Users/lukasz/go/src/github.com/lukaszbudnik/migrator/test/migrations/tenants/201602160002.sql",
    "migrationType": 2,
    "contents": "create table {schema}.module (id integer, id_config integer, foreign key (id_config) references config.config(id));\n",
    "checkSum": "56c4c1d8f82f3dedade5116be46267edee01a4889c6359ef03c39dc73ca653a8",
    "schema": "abc",
    "appliedAt": "2020-01-01T17:29:13.169306Z"
  },
  {
    "name": "201602160002.sql",
    "sourceDir": "/Users/lukasz/go/src/github.com/lukaszbudnik/migrator/test/migrations/tenants",
    "file": "/Users/lukasz/go/src/github.com/lukaszbudnik/migrator/test/migrations/tenants/201602160002.sql",
    "migrationType": 2,
    "contents": "create table {schema}.module (id integer, id_config integer, foreign key (id_config) references config.config(id));\n",
    "checkSum": "56c4c1d8f82f3dedade5116be46267edee01a4889c6359ef03c39dc73ca653a8",
    "schema": "def",
    "appliedAt": "2020-01-01T17:29:13.169306Z"
  },
  {
    "name": "201602160002.sql",
    "sourceDir": "/Users/lukasz/go/src/github.com/lukaszbudnik/migrator/test/migrations/tenants",
    "file": "/Users/lukasz/go/src/github.com/lukaszbudnik/migrator/test/migrations/tenants/201602160002.sql",
    "migrationType": 2,
    "contents": "create table {schema}.module (id integer, id_config integer, foreign key (id_config) references config.config(id));\n",
    "checkSum": "56c4c1d8f82f3dedade5116be46267edee01a4889c6359ef03c39dc73ca653a8",
    "schema": "xyz",
    "appliedAt": "2020-01-01T17:29:13.169306Z"
  }
]
```

`MigrationDB` JSON contains all the fields from `Migration` struct and adds the following ones:

* `schema` - schema for which given migration was applied, for single migrations this is equal to source dir name, for multi-tenant ones this is the name of the actual tenant schema
* `appliedAt` - date time migration was applied


## POST /v1/migrations

Applies new source migrations to DB and returns summary results and a list of applied migrations.

This operation requires as an input the following JSON payload:

* `mode` - defines mode in which migrator will execute migrations, valid values are:
  * `apply` - applies migrations
  * `sync` - synchronises all source migrations with internal migrator's table, this action loads and marks all source migrations as applied but does not apply them
  * `dry-run` - instead of calling commit, calls rollback at the end of the operation
* `response` - controls how much information is returned by migrator, valid values are:
  * `full` - the response will contain both summary results and a list of applied migrations/scripts
  * `list` - the response will contain both summary results and a list of applied migrations/scripts but without their contents (introduced in migrator `v4.1.1` and a part of API v1; does not break API v1 contract - existing integrations will continue to work)
  * `summary` - the response will contain only summary results

Sample request:

```
curl -v -X POST -H "Content-Type: application/json" -d '{"mode": "apply", "response": "full"}' http://localhost:8080/v1/migrations
```

Sample HTTP response:

```
{
  "results": {
    "startedAt": "2020-01-01T18:29:13.14682+01:00",
    "duration": 51637303,
    "tenants": 3,
    "singleMigrations": 4,
    "tenantMigrations": 4,
    "tenantMigrationsTotal": 12,
    "migrationsGrandTotal": 16,
    "singleScripts": 1,
    "tenantScripts": 1,
    "tenantScriptsTotal": 3,
    "scriptsGrandTotal": 4
  },
  "appliedMigrations": [
    {
      "name": "201602160001.sql",
      "sourceDir": "/Users/lukasz/go/src/github.com/lukaszbudnik/migrator/test/migrations/config",
      "file": "/Users/lukasz/go/src/github.com/lukaszbudnik/migrator/test/migrations/config/201602160001.sql",
      "migrationType": 1,
      "contents": "create schema config;\n",
      "checkSum": "c1380af7a054ec75778252f539e1e9f914d2c5b1f441ea1df18c2140c6c3380a"
    }
  ]
}
```

`appliedMigrations` is a list of JSON representation of `Migration` struct as already described above.

`results` is a JSON representation of `Results` struct. `Results` JSON contains the following fields:

* `startedAt` - date time the operation started
* `duration` - how long the operation took in nanoseconds
* `tenants` - number of tenants in the system
* `singleMigrations` - number of identified and applied single migrations
* `tenantMigrations` - number of identified tenant migrations
* `tenantMigrationsTotal` - number of all tenant migrations applied (equals to `tenants` * `tenantMigrations`)
* `migrationsGrandTotal` - sum of `singleMigrations` and `tenantMigrationsTotal`
* `singleScripts` - number of read and applied single scripts
* `tenantScripts` - number of read tenant scripts
* `tenantScriptsTotal` - number of all tenant scripts applied (equals to `tenants` * `tenantScripts`)
* `scriptsGrandTotal` - sum of `singleScripts` and `tenantScriptsTotal`


## GET /v1/tenants

Returns list of all tenants.

Sample request:

```
curl -v http://localhost:8080/v1/tenants
```

Sample HTTP response:


```
< HTTP/1.1 200 OK
< Content-Type: application/json; charset=utf-8
< Date: Wed, 01 Jan 2020 17:16:09 GMT
< Content-Length: 58

["abc","def","xyz","new_test_tenant_1577793069634018000"]
```

## POST /v1/tenants

Adds a new tenant and applies all tenant migrations and scripts for newly created tenant. Returns summary results and a list of applied migrations.

This operation requires as an input the following JSON payload:

* `name` - the name of the new tenant
* `mode` - same as `mode` for [POST /v1/migrations](#post-v1migrations)
* `response` - same as `response` for [POST /v1/migrations](#post-v1migrations)

Sample request:

```
curl -v -X POST -H "Content-Type: application/json" -d '{"name": "new_test_tenant", "mode": "apply", "response": "full"}' http://localhost:8080/v1/tenants
```

Sample HTTP response.

```
< HTTP/1.1 200 OK
< Content-Type: application/json; charset=utf-8
< Date: Wed, 01 Jan 2020 17:45:00 GMT
< Transfer-Encoding: chunked

{
  "results": {
    "startedAt": "2020-01-01T18:45:00.174152+01:00",
    "duration": 12426788,
    "tenants": 1,
    "singleMigrations": 0,
    "tenantMigrations": 4,
    "tenantMigrationsTotal": 4,
    "migrationsGrandTotal": 4,
    "singleScripts": 0,
    "tenantScripts": 1,
    "tenantScriptsTotal": 1,
    "scriptsGrandTotal": 1
  },
  "appliedMigrations": [
    {
      "name": "201602160002.sql",
      "sourceDir": "/Users/lukasz/go/src/github.com/lukaszbudnik/migrator/test/migrations/tenants",
      "file": "/Users/lukasz/go/src/github.com/lukaszbudnik/migrator/test/migrations/tenants/201602160002.sql",
      "migrationType": 2,
      "contents": "create table {schema}.module (id integer, id_config integer, foreign key (id_config) references config.config(id));\n",
      "checkSum": "56c4c1d8f82f3dedade5116be46267edee01a4889c6359ef03c39dc73ca653a8"
    }
  ]
}
```

The response is identical to the one of [POST /v1/migrations](#post-v1migrations). When adding new tenant only tenant migrations and scripts are applied and only for the newly created tenant. That is why `singleMigrations` and `singleScripts` are always 0 and `tenants` is always 1.

## Request tracing

migrator uses request tracing via `X-Request-ID` header. This header can be used with all requests for tracing and/or auditing purposes. If this header is absent migrator will generate one for you.

# Quick Start Guide

You can apply your first migrations with migrator in literally a couple of minutes. There are some test migrations which are located in `test` directory as well as some docker scripts for setting up test databases.

The quick start guide shows you how to either build the migrator locally or use the official docker image.

Steps 1 & 2 are required either way (migrator source code contains sample configuration & setup files together with some test migrations).
Step 3 is for building migrator locally, step 4 is for running the migrator container.
Step 5 is running examples and enjoying migrator ;)

## 1. Get the migrator project

Get the source code the usual go way:

```
go get -d -v github.com/lukaszbudnik/migrator
cd $GOPATH/src/github.com/lukaszbudnik/migrator
```

migrator aims to support 3 latest Go versions (built automatically on Travis).

## 2. Setup test DB container

migrator comes with helper scripts to setup test DB containers. Let's use postgres (see `ultimate-coverage.sh` for all supported containers).

```
./test/docker/create-and-setup-container.sh postgres
```

Script will start container called `migrator-postgres`.

Further, apart of starting test DB container, the script also generates a ready-to-use test config file. We will use it later.

## 3. Build and run migrator

migrator uses go modules to manage dependencies. When building & running migrator from source code simply execute:

```
go build
./migrator -configFile test/migrator.yaml
```

> Note: There are 3 git variables injected into the production build (branch/tag together with commit sha & commit date). When migrator is built like above it prints empty branch/tag, commit sha and date. This is OK for local development. If you want to inject proper values take a look at `Dockerfile` for details.

## 4. Run migrator from official docker image

The official migrator docker image is available on docker hub [lukasz/migrator](https://hub.docker.com/r/lukasz/migrator).

All migrator releases are automatically available as docker images on docker hub [lukasz/migrator/tags](https://hub.docker.com/r/lukasz/migrator/tags).

To start a migrator container you need to:

1. mount a volume with migrations, for example: `/data`
2. specify location of migrator configuration file, for convenience it is usually located under `/data` directory; it defaults to `/data/migrator.yaml` and can be overridden by setting environment variable `MIGRATOR_YAML`

When running migrator from docker we need to update `migrator.yaml` (generated in step 2) as well as provide a link to `migrator-postgres` container:

```
sed -i "s/host=[^ ]* port=[^ ]*/host=migrator-postgres port=5432/g" test/migrator.yaml
sed -i "s/baseLocation: .*/baseLocation: \/data\/migrations/g" test/migrator.yaml
docker run --name migrator-test -p 8080:8080 -v $PWD/test:/data -e MIGRATOR_YAML=/data/migrator.yaml -d --link migrator-postgres lukasz/migrator
```

## 5. Play around with migrator

Happy path:

```
curl -v http://localhost:8080/v1/config
curl -v http://localhost:8080/v1/migrations/source
curl -v http://localhost:8080/v1/tenants
curl -v http://localhost:8080/v1/migrations/applied
curl -v -X POST -H "Content-Type: application/json" -d '{"mode": "apply", "response": "full"}' http://localhost:8080/v1/migrations
curl -v -X POST -H "Content-Type: application/json" -d '{"name": "new_tenant", "mode": "apply", "response": "full"}' http://localhost:8080/v1/tenants
curl -v http://localhost:8080/v1/migrations/applied
```

And some errors. For example let's break a checksum of the first migration and try to apply migrations or add new tenant.

```
echo " " >> test/migrations/config/201602160001.sql
curl -v -X POST -H "Content-Type: application/json" -d '{"mode": "apply", "response": "full"}' http://localhost:8080/v1/migrations
curl -v -X POST -H "Content-Type: application/json" -d '{"name": "new_tenant", "mode": "apply", "response": "full"}' http://localhost:8080/v1/tenants
```

# Tutorials

In this section I provide links to more in-depth migrator tutorials.

## Deploying migrator to AWS ECS

The goal of this tutorial is to deploy migrator to AWS ECS, load migrations from AWS S3 and apply them to AWS RDS DB while storing env variables securely in AWS Secrets Manager. The list of all AWS services used is: IAM, ECS, ECR, Secrets Manager, RDS, and S3.

You can find it in [contrib/aws-ecs-ecr-secretsmanager-rds-s3](contrib/aws-ecs-ecr-secretsmanager-rds-s3).

## Deploying migrator to AWS EKS

The goal of this tutorial is to deploy migrator to AWS EKS, load migrations from AWS S3 and apply them to AWS RDS DB. The list of AWS services used is: IAM, EKS, ECR, RDS, and S3.

You can find it in [contrib/kubernetes-aws-eks](contrib/kubernetes-aws-eks).

## Deploying migrator to Azure AKS

The goal of this tutorial is to publish migrator image to Azure ACR private container repository, deploy migrator to Azure AKS, load migrations from Azure Blob Container and apply them to Azure Database for PostgreSQL. The list of Azure services used is: AKS, ACR, Blob Storage, and Azure Database for PostgreSQL.

You can find it in [contrib/azure-aks](contrib/azure-aks).

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
# URL and template are required if at least one of them is empty noop notifier is used
# the default content type header sent is application/json (can be overridden via webHookHeaders below)
webHookURL: https://your.server.com/services/TTT/BBB/XXX
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

## Source migrations

Migrations can be read either from local disk or from S3 (I'm open to contributions to add more cloud storage options).

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
* Microsoft SQL Server 2017 - a relational database management system developed by Microsoft, driver used: https://github.com/denisenkom/go-mssqldb
  * Microsoft SQL Server

# Customisation and legacy frameworks support

migrator can be used with an already existing legacy DB migration framework.

## Custom tenants support

If you have an existing way of storing information about your tenants you can configure migrator to use it.
In the config file you need to provide 2 configuration properties:

* `tenantSelectSQL` - a select statement which returns names of the tenants
* `tenantInsertSQL` - an insert statement which creates a new tenant entry, the insert statement should be a valid prepared statement for the SQL driver/database you use, it must accept the name of the new tenant as a parameter; finally should your table require additional columns you need to provide default values for them too

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

Before switching from a legacy framework to migrator you need to synchronise source migrations to migrator.

This can be done using the POST /v1/migrations endpoint and setting the `mode` param to `sync`:

```
curl -v -X POST -H "Content-Type: application/json" -d '{"mode": "sync", "response": "full"}' http://localhost:8080/v1/migrations
```

migrator will load and synchronise all source migrations with internal migrator's table, this action loads and marks all source migrations as applied but does not apply them.

Once the initial sync is done you can move to migrator for all the consecutive DB migrations.

## Final comments

When using migrator please remember that:

* migrator creates `migrator` schema together with `migrator_migrations` table automatically
* if you're not using [Custom tenants support](#custom-tenants-support) migrator creates `migrator_tenants` table automatically; just like `migrator_migrations` this table is created inside the `migrator` schema
* when adding a new tenant migrator creates a new DB schema and applies all tenant migrations and scripts - no need to apply them manually
* single schemas are not created automatically, you must add initial migration with `create schema {schema}` SQL statement (see examples above)

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
./ultimate-coverage.sh
```

The `ultimate-coverage.sh` script loops through 5 different containers (3 MySQL flavours, PostgreSQL, and MSSQL) creates db docker container, executes `coverage.sh` script, and finally tears down given db docker container.

# License

Copyright 2016-2020 Łukasz Budnik

Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except in compliance with the License. You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the specific language governing permissions and limitations under the License.
