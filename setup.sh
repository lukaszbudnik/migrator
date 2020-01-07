#!/usr/bin/env sh

# this is for dockerhub failing on fetching packages from gopkg.in
# travis resolves this by 3 retries so adapting 3 retries here as well
for i in {1..3}; do
  # explicit gin get makes sure validator package is fetched before other packages are fetched
  go get -v github.com/gin-gonic/gin
  go get -t -v ./... && break || sleep 15;
done
