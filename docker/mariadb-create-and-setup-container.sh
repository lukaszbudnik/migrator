#!/bin/bash

set -x

cd `dirname $0`

ip=`docker-machine ip default`
port=33066
database='migrator_test'

docker run -d \
  --name migrator-mariadb \
  -e MYSQL_ALLOW_EMPTY_PASSWORD=yes \
  -p $port:3306 \
  mariadb

sleep 10

running=`docker inspect -f {{.State.Running}} migrator-mariadb`

if [[ "true" == "$running" ]]; then

  sleep 10

  mysql -u root -h $ip -P $port -e "create database $database"
  mysql -u root -h $ip -P $port -D $database < ../test/create-test-tenants.sql

  cat ../test/migrator-mysql.yaml | sed "s/\"[^@]*@[^:]*:[^\/]*\/[^\"]*/\"root@($ip:$port)\/$database/g" > ../test/migrator.yaml
else
  echo "Could not setup mariadb"
fi
