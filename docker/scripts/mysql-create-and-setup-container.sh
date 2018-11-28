#!/bin/bash

function mysql_start() {

  flavour=$1

  ip=127.0.0.1
  port=33306
  database='migrator_test'

  docker run -d \
    --name "migrator-$flavour" \
    -e MYSQL_ALLOW_EMPTY_PASSWORD=yes \
    -p $port:3306 \
    "$flavour"

  # mysql needs a little bit more time
  sleep 20

  running=$(docker inspect -f {{.State.Running}} "migrator-$flavour")

  if [[ "true" == "$running" ]]; then
    mysql -u root -h $ip -P $port -e "create database $database"
    mysql -u root -h $ip -P $port -D $database < ../test/create-test-tenants.sql
    cat ../test/migrator-mysql.yaml | sed "s/A/root:@tcp($ip:$port)\/$database?parseTime=true/g" > ../test/migrator.yaml
  else
    >&2 echo "Could not setup mysql-$flavour"
    exit 1
  fi

}
