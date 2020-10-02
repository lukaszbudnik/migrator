#!/bin/bash

# assumes containers are already started
# it uses PostgreSQL and requires psql tool to be installed

NO_OF_MIGRATIONS=100

go build

# generate test migrtations
./test/performance/generate-test-migrations.sh -n $NO_OF_MIGRATIONS

export PGPASSWORD=supersecret
psql -U postgres -h 127.0.0.1 -p 5432 -d migrator -tAq -c "delete from migrator.migrator_versions"

./migrator -configFile=test/performance/migrator-performance.yaml  &
PID=$!
sleep 5

COMMIT_SHA="performance-tests"
# new lines are used for readability but have to be removed from the actual request
cat <<EOF | tr -d "\n" > create_version.txt
{
  "query": "
  mutation CreateVersion(\$input: VersionInput!) {
    createVersion(input: \$input) {
      version {
        id,
        name,
      }
      summary {
        startedAt
        tenants
        migrationsGrandTotal
        scriptsGrandTotal
      }
    }
  }",
  "operationName": "CreateVersion",
  "variables": {
    "input": {
      "versionName": "$COMMIT_SHA"
    }
  }
}
EOF
# and now execute the above query
start=`date +%s`
curl -d @create_version.txt http://localhost:8888/v2/service
end=`date +%s`

output=$(psql -U postgres -h 127.0.0.1 -p 5432 -d migrator -tAq -c "select name from migrator.migrator_tenants")

IFS=$'\n'; tenants=($output); unset IFS;

for tenant in "${tenants[@]}"; do
    count=$(psql -U postgres -h 127.0.0.1 -p 5432 -d migrator -tAq -c "select count(distinct col) from $tenant.table_for_inserts")
    tables_count=$(psql -U postgres -h 127.0.0.1 -p 5432 -d migrator -tAq -c "select count(*) from information_schema.tables where table_schema = '$tenant' and table_name like 'table_%';")

    if [[ $count -ne $NO_OF_MIGRATIONS+1 ]]; then
      echo "[migrations inserts] error for $tenant, got $count, expected: $((NO_OF_MIGRATIONS+1))"
    fi
    if [[ $tables_count -ne $NO_OF_MIGRATIONS/10+1 ]]; then
      echo "[migrations tables] error for $tenant, got $tables_count, expected: $((NO_OF_MIGRATIONS/10+1))"
    fi

done

echo "Test took $((end-start)) seconds"

# remove generated migrations and test request and kill migrator
rm -rf test/performance/migrations
rm create_version.txt
kill $PID
