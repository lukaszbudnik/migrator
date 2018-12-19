#!/bin/sh

DEFAULT_YAML_LOCATION="/data/migrator.yaml"

# if migrator config file is not provided explicitly fallback to default location
if [ -z "$MIGRATOR_YAML" ]; then
  MIGRATOR_YAML=$DEFAULT_YAML_LOCATION
fi

migrator -configFile "$MIGRATOR_YAML"
