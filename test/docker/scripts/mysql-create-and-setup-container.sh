#!/bin/bash

function mysql_start() {
  flavour=$1
  name=${flavour//\//-}

  ip=127.0.0.1
  port=3306
  database=migrator_test

  docker run -d \
    --name "migrator-$flavour" \
    -e MYSQL_ALLOW_EMPTY_PASSWORD=yes \
    -P \
    "$flavour"

  sleep 30

  running=$(docker inspect -f {{.State.Running}} "migrator-$name")

  if [[ "true" == "$running" ]]; then
    docker_port=$(docker port "migrator-$name" | grep "^$port/tcp" | awk -F ':' '{print $2}')

    mysql -u root -h $ip -P $docker_port -e "create database $database"
    mysql -u root -h $ip -P $docker_port -D $database < ../create-test-tenants.sql
    cat ../migrator-mysql.yaml | sed "s/A/root:@tcp($ip:$docker_port)\/$database?parseTime=true\&timeout=1s/g" > ../migrator.yaml
  else
    >&2 echo "Could not setup mysql-$name"
    exit 1
  fi

}
