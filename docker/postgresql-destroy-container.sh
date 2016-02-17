#!/bin/bash

set -x

docker stop migrator-postgresql

docker rm migrator-postgresql
