#!/bin/bash

NO_OF_MIGRATIONS=1000
NO_OF_TENANTS=10

export PGPASSWORD=supersecret
# remove all existing tenants
psql -U postgres -h 127.0.0.1 -p 5432 -d migrator -tAq -c "delete from migrator.migrator_tenants"
# create test tenants (connects to psql and creates them)
./test/performance/create-test-tenants.sh $NO_OF_TENANTS

./test/performance/generate-test-migrations.sh -f -n $NO_OF_MIGRATIONS

# flyway doesn't support both single and multi-tenant migrations, delete public ones
rm -rf ./test/performance/migrations/public

# remove existing flyway.schemas config property
gsed -i '/flyway.schemas/d' ./test/performance/flyway.conf

output=$(psql -U postgres -h 127.0.0.1 -p 5432 -d migrator -tAq -c "select string_agg(name, ',') from migrator.migrator_tenants")

echo "flyway.schemas=$output" >> ./test/performance/flyway.conf

start=`date +%s`
flyway -configFiles=./test/performance/flyway.conf baseline migrate > /dev/null
end=`date +%s`

echo "Test took $((end-start)) seconds"

# append test
# ./test/performance/generate-test-migrations.sh -a -f -n $NO_OF_MIGRATIONS
# start=`date +%s` && flyway -configFiles=./test/performance/flyway.conf migrate > /dev/null && end=`date +%s`
