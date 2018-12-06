#!/bin/bash

# when called with no arguments calls tests for all packages
if [[ -z "$1" ]]; then
  packages=$(go list -f "{{.Name}}" ./...)
else
  packages="$1"
fi

echo "mode: set" > coverage-all.txt

go clean -testcache

for package in $packages
do
  if [[ "main" == "$package" ]]; then
    continue
  fi
  go test -cover -coverprofile=coverage-$package.txt ./$package
  cat coverage-$package.txt | sed '/^mode/d' >> coverage-all.txt
done
