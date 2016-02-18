#!/bin/bash

set -x

docker stop migrator-mariadb

docker rm migrator-mariadb
