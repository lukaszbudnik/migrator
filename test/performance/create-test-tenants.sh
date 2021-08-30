#!/bin/bash

# not the most fancy script in the world

if [[ $# -eq 0 ]]; then
  echo "This script expects a number of test tenants to create as its argument"
  exit
fi

i=1
let end=$1+1

while [[ $i -lt $end ]]; do
  if [[ $i%10 -eq 0 ]]; then
    echo "creating new tenant $i"
  fi
  name="tenant_${RANDOM}_${RANDOM}"
  psql -U postgres -h 127.0.0.1 -p 5432 -d migrator -tAq -c "create schema $name; insert into migrator.migrator_tenants (name) values ('$name');"
  let i+=1
done
