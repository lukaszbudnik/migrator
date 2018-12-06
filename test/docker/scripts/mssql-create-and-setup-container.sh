#!/bin/bash

function mssql_start() {
  name=mssql

  ip=127.0.0.1
  port=1433
  database=migratortest
  password=YourStrongPassw0rd

  docker run -d \
     --name "migrator-$name" \
     -e 'ACCEPT_EULA=Y' -e "SA_PASSWORD=$password" \
     -P \
     mcr.microsoft.com/mssql/server:2017-latest

  sleep 15

  running=$(docker inspect -f {{.State.Running}} "migrator-$name")

  if [[ "true" == "$running" ]]; then
    docker_port=$(docker port "migrator-$name" | grep "^$port/tcp" | awk -F ':' '{print $2}')

    sqlcmd -S "$ip,$docker_port" -U SA -P $password -Q "CREATE DATABASE $database"
    sqlcmd -S "$ip,$docker_port" -U SA -P $password -d $database -i ../create-test-tenants-mssql.sql

    cat ../migrator-mssql.yaml | sed "s/A/sqlserver:\/\/SA:$password@$ip:$docker_port\/?database=$database\&connection+timeout=1\&dial+timeout=1/g" > ../migrator.yaml
  else
    >&2 echo "Could not setup mssql-$name"
    exit 1
  fi
}
