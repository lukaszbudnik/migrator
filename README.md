# Migrator [![Build Status](https://travis-ci.org/lukaszbudnik/migrator.svg?branch=master)](https://travis-ci.org/lukaszbudnik/migrator)

DB migration tool written in go.

# Usage

Short and sweet. Migrator requires a simple `migrator.yaml` file:

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

Migrator will scan all directories under `baseDir` directory. Migrations listed under `singleSchemas` will be applied once. Migrations listed under `tenantSchemas` will be applied for all tenants fetched using `tenantsSql`.

# Supported databases

Currently migrator supports the following databases:

* PostgreSQL - true multitenant database, with transactions spanning DDL statements
* MySQL - pseudo multitenant database, transactions do not span DDL statements
* MariaDB - enhanced MySQL

# Examples

PostgreSQL:

```
$ docker/postgresql-create-and-setup-container.sh
$ go test -v
$ docker/postgresql-destroy-container.sh
```

MySQL:

```
$ docker/mysql-create-and-setup-container.sh
$ go test -v
$ docker/mysql-destroy-container.sh
```

MariaDB:

```
$ docker/mariadb-create-and-setup-container.sh
$ go test -v
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
