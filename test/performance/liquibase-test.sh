#!/bin/bash

# liquibase doesn't natively support multi-tenancy thus doing a single migration to public schema
NO_OF_MIGRATIONS=10000

./test/performance/generate-test-migrations.sh -f -n $NO_OF_MIGRATIONS

# download PostgreSQL JDBC Driver - update location in liquibase.properties!
mvn -DgroupId=org.postgresql -DartifactId=postgresql -Dversion=42.2.23 dependency:get

start=`date +%s`
liquibase --defaultsFile ./test/performance/liquibase.properties update
end=`date +%s`

echo "Test took $((end-start)) seconds"

rm -rf test/performance/migrations/

# append test
# 1. comment out above rm command
# 2. RUN TEST
# 3. generate new migrations:
# ./test/performance/generate-test-migrations.sh -a -f -n $NO_OF_MIGRATIONS
# 4. execute liquibase migrate update, measure start and end times:
# start=`date +%s` && liquibase --defaultsFile ./test/performance/liquibase.properties update && end=`date +%s`
