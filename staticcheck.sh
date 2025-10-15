#!/bin/bash

which staticcheck &> /dev/null
if [[ $? -ne 0 ]]; then
    go install honnef.co/go/tools/cmd/staticcheck@latest
fi

staticcheck ./...
