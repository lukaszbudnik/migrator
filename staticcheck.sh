#!/bin/bash

which staticcheck &> /dev/null
if [[ $? -ne 0 ]]; then
    go install honnef.co/go/tools/cmd/staticcheck@latest
fi

# migrator didn't receive major update since go 1.17 there are a couple of deprecated API used there

staticcheck ./... || true
