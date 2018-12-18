#!/bin/sh

DEFAULT_YAML_LOCATION="/data/migrator.yaml"

# if migrator config file is not provided explicitly fallback to default location
if [ -z "$MIGRATOR_YAML" ]; then
  MIGRATOR_YAML=$DEFAULT_YAML_LOCATION
fi

if [ ! -f "$MIGRATOR_YAML" ] ; then
  echo "Migrator config file not found. Please use either default location ($DEFAULT_YAML_LOCATION) or provide a custom one using \$MIGRATOR_YAML."
  exit 1
fi

echo "Starting migrator using config file: $MIGRATOR_YAML"
migrator -configFile "$MIGRATOR_YAML"
