#!/bin/bash

docker run -d \
  --name migrator-postgresql \
  -p 55432:5432 \
  postgres
