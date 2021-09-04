#!/bin/bash

go install honnef.co/go/tools/cmd/staticcheck@latest

staticcheck ./...
