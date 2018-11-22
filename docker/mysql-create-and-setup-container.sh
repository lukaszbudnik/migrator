#!/bin/bash

set -x

cd `dirname $0`

ip=127.0.0.1
port=33306
database='migrator_test'

docker run -d \
  --name migrator-mysql \
  -e MYSQL_ALLOW_EMPTY_PASSWORD=yes \
  -p $port:3306 \
  mysql

# mysql needs a little bit more time
sleep 20

running=`docker inspect -f {{.State.Running}} migrator-mysql`

if [[ "true" == "$running" ]]; then
  mysql -u root -h $ip -P $port -e "create database $database"
  mysql -u root -h $ip -P $port -D $database < ../test/create-test-tenants.sql
  cat ../test/migrator-mysql.yaml | sed "s/A/root:@tcp($ip:$port)\/$database?parseTime=true/g" > ../test/migrator.yaml
else
  echo "Could not setup mysql"
fi
