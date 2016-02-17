#!/bin/bash

set -x

docker stop migrator-mysql

docker rm migrator-mysql
