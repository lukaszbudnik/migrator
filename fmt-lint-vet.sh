#!/bin/bash

gofmt -s -w .
golint ./...

packages=`go list -f "{{.Name}}" ./...`

for package in $packages
do
  if [[ "main" == "$package" ]]; then
    continue
  fi
  go vet ./$package
done
