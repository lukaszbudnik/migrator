# Contributing, code style, running unit & integration tests

Contributions are most welcomed.

If you would like to help me and implement a new feature, enhance existing one, or spotted and fixed bug please send me a pull request.

Code should be formatted, checked, and tested using the following commands:

```
docker-compose -f test/docker-compose.yaml up
./fmt-lint-vet.sh
./coverage.sh
./test/http-integration-tests.sh
```

The `db/db_integration_test.go` uses go subtests and runs all tests agains 5 different database containers (3 MySQL flavours, PostgreSQL, and MSSQL). These databases are automatically provisioned by the docker-compose tool.
