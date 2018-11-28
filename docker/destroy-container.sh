#!/usr/bin/env bash

set -x

cd `dirname $0`

CONTAINER_TYPE=$1

for SCRIPT in scripts/*.sh; do
  source "${SCRIPT}"
done

destroy_container $CONTAINER_TYPE
