#!/bin/bash

# when called with no arguments calls tests for all packages
if [[ -z "$1" ]]; then
  packages=$(go list -f "{{.Name}}" ./...)
else
  packages="$1"
fi

echo "mode: atomic" > coverage.txt

go clean -testcache

fail=0

for package in $packages
do
  if [[ "main" == "$package" ]]; then
    continue
  fi
  go test -race -covermode=atomic -coverprofile=coverage-$package.txt ./$package
  if [[ $? -ne 0 ]]; then
    fail=1
  fi
  cat coverage-$package.txt | sed '/^mode/d' | sed '/_mocks.go/d' >> coverage.txt
  rm coverage-$package.txt
done

exit $fail
