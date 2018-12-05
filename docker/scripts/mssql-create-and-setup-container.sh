#!/bin/bash

function mssql_start() {
  name=mssql

  ip=127.0.0.1
  port=11433
  database=migratortest
  password=YourStrongPassw0rd

  docker run -e 'ACCEPT_EULA=Y' -e "SA_PASSWORD=$password" \
     -p $port:1433 --name "migrator-$name" \
     -d mcr.microsoft.com/mssql/server:2017-latest

  sleep 10

  running=$(docker inspect -f {{.State.Running}} "migrator-$name")

  if [[ "true" == "$running" ]]; then
    sqlcmd -S "$ip,$port" -U SA -P $password -Q "CREATE DATABASE $database"
    sqlcmd -S "$ip,$port" -U SA -P $password -d $database -i ../test/create-test-tenants-mssql.sql

    cat ../test/migrator-mssql.yaml | sed "s/A/sqlserver:\/\/SA:$password@$ip:$port\/?database=$database/g" > ../test/migrator.yaml
  else
    >&2 echo "Could not setup mssql-$name"
    exit 1
  fi
}
