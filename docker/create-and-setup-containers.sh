#!/bin/bash

set -x

cd `dirname $0`

ip=`docker-machine ip default`
port=55432

docker run -d \
  --name migrator-postgresql \
  -p $port:5432 \
  postgres

sleep 10

running=`docker inspect -f {{.State.Running}} migrator-postgresql`

if [[ "true" == "$running" ]]; then
  psql -U postgres -c 'create database "migrator-test"' -h $ip -p $port
  psql -U postgres -f ../test/create-test-tenants.sql -h $ip -p $port -d migrator-test

  cat ../test/migrator-test.yaml | sed "s/host=[^ ]* port=[^ ]*/host=$ip port=$port/g" > ../test/migrator.yaml
else
  echo "Could not setup migrator-test db"
fi
