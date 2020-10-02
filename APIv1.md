## /v1

**As of migrator v2020.1.0 API v1 is deprecated and will sunset in v2021.1.0.**

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