#!/bin/bash

which gotestsum &> /dev/null
if [[ $? -ne 0 ]]; then
  go install gotest.tools/gotestsum@latest
fi

# when called with no arguments calls tests for all packages
if [[ -z "$1" ]]; then
  packages='./...'
else
  packages="./$1"
fi

fail=0

gotestsum --junitfile unit-tests.xml -- -covermode=atomic -coverprofile=coverage-all.txt $packages

if [[ $? -ne 0 ]]; then
  fail=1
fi

cat coverage-all.txt | sed '/_mocks.go/d' > coverage.txt

rm coverage-all.txt

exit $fail
