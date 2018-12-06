#!/bin/bash

dbs="postgres mysql mariadb percona mssql"

for db in $dbs
do
  echo "1. destroy $db docker (just in case it's already created)"
  ./test/docker/destroy-container.sh $db
  echo "2. create and setup $db docker"
  ./test/docker/create-and-setup-container.sh $db
  echo "3. run all tests"
  ./coverage.sh
  echo "4. destroy $db docker (cleanup)"
  ./test/docker/destroy-container.sh $db
done
