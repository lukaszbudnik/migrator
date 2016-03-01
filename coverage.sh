#!/bin/bash

if [[ -z $1 ]]; then
  packages=`go list -f "{{.Name}}" ./...`
else
  packages=$1
fi

for package in $packages
do
  if [[ "main" == "$package" ]]; then
    continue
  fi
  go test -cover -coverprofile=coverage-$package.txt ./$package
done
