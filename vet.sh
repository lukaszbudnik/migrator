#!/bin/bash

# when called with no arguments calls tests for all packages
if [[ -z "$1" ]]; then
  packages=`go list -f "{{.Name}}" ./...`
else
  packages="$1"
fi

for package in $packages
do
  if [[ "main" == "$package" ]]; then
    continue
  fi
  go vet ./$package
done
