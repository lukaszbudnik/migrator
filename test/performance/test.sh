#!/bin/bash

NO_OF_MIGRATIONS=100

go build

# create test container
./docker/postgresql-create-and-setup-container.sh

cp test/migrator.yaml test/performance

cd test/performance
cat migrator.yaml
sed -i "s/baseDir: [^ ]*/baseDir: migrations/g" migrator.yaml
echo 'schemaPlaceHolder: ":tenant"' >> migrator.yaml

# generate test migrtations
./generate-test-migrations.sh $NO_OF_MIGRATIONS

cat migrator.yaml

psql -U postgres -h 127.0.0.1 -p 55432 -d migrator_test -tAq -c "delete from public.migrator_migrations"

start=`date +%s`
migrator
end=`date +%s`

output=$(psql -U postgres -h 127.0.0.1 -p 55432 -d migrator_test -tAq -c "select name from public.migrator_tenants")

IFS=$'\n'; tenants=($output); unset IFS;

for tenant in "${tenants[@]}"; do
    count=$(psql -U postgres -h 127.0.0.1 -p 55432 -d migrator_test -tAq -c "select count(distinct col) from $tenant.table_for_inserts")
    tables_count=$(psql -U postgres -h 127.0.0.1 -p 55432 -d migrator_test -tAq -c "select count(*) from information_schema.tables where table_schema = '$tenant' and table_name like 'table_%';")

    if [[ $count -ne $NO_OF_MIGRATIONS+1 ]]; then
      echo "[migrations inserts] error for $tenant, got $count, expected: $((NO_OF_MIGRATIONS+1))"
    fi
    if [[ $tables_count -ne $NO_OF_MIGRATIONS/10+1 ]]; then
      echo "[migrations tables] error for $tenant, got $tables_count, expected: $((NO_OF_MIGRATIONS/10+1))"
    fi

done

echo "Test took $((end-start)) seconds"

# remove generated migrations and test config
rm -rf migrations
rm migrator.yaml

# stop and destroy the test container
cd ../..
./docker/postgresql-destroy-container.sh
