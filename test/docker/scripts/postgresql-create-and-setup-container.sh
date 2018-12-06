#!/bin/bash

function postgresql_start() {
  flavour=$1
  name=${flavour//\//-}

  ip=127.0.0.1
  port=5432
  database=migrator_test

  docker run -d \
    --name "migrator-$flavour" \
    -P \
    "$flavour"

  sleep 10

  running=$(docker inspect -f {{.State.Running}} "migrator-$name")

  if [[ "true" == "$running" ]]; then
    docker_port=$(docker port "migrator-$name" | grep "^$port/tcp" | awk -F ':' '{print $2}')

    psql -U postgres -h $ip -p $docker_port -c "create database $database"
    psql -U postgres -h $ip -p $docker_port -d $database -f ../create-test-tenants.sql

    cat ../migrator-postgresql.yaml | sed "s/dbname=[^ ]* host=[^ ]* port=[^ ]*/dbname=$database host=$ip port=$docker_port/g" > ../migrator.yaml
  else
    >&2 echo "Could not setup postgresql-$name"
    exit 1
  fi
}
