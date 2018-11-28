#!/bin/bash

ip=127.0.0.1
port=55432
database=migrator_test

function postgresql_start() {
  docker run -d \
    --name migrator-postgresql \
    -p $port:5432 \
    postgres

  sleep 10

  running=`docker inspect -f {{.State.Running}} migrator-postgresql`

  if [[ "true" == "$running" ]]; then
    psql -U postgres -h $ip -p $port -c "create database $database"
    psql -U postgres -h $ip -p $port -d $database -f ../test/create-test-tenants.sql

    cat ../test/migrator-postgresql.yaml | sed "s/dbname=[^ ]* host=[^ ]* port=[^ ]*/dbname=$database host=$ip port=$port/g" > ../test/migrator.yaml
  else
    >&2 echo "Could not setup postgresql"
    exit 1
  fi
}
