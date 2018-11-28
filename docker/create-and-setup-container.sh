#!/usr/bin/env bash

set -x

cd `dirname $0`

CONTAINER_TYPE=$1

for SCRIPT in scripts/*.sh; do
  source "$SCRIPT"
done

case $CONTAINER_TYPE in
  postgresql )
    postgresql_start
    ;;
  mysql )
    mysql_start mysql
    ;;
  mariadb )
    mysql_start mariadb
    ;;
  * )
    >&2 echo "Unknown container type $CONTAINER_TYPE"
    exit 1

esac
