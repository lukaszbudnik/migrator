#!/bin/bash

# assumes containers are already started
# it uses PostgreSQL and requires psql tool to be installed

NO_OF_MIGRATIONS=100
NO_OF_TENANTS=500

# test from scratch
EXISTING_TABLES=0
EXISTING_INSERTS=0
# in append test
# EXISTING_TABLES=10
# EXISTING_INSERTS=100

go build

export PGPASSWORD=supersecret
# remove all existing versions and migrations
psql -U postgres -h 127.0.0.1 -p 5432 -d migrator -tAq -c "delete from migrator.migrator_versions"
# remove all existing tenants
psql -U postgres -h 127.0.0.1 -p 5432 -d migrator -tAq -c "delete from migrator.migrator_tenants"
# create test tenants (connects to psql and creates them)
./test/performance/create-test-tenants.sh $NO_OF_TENANTS

# generate test migrtations
./test/performance/generate-test-migrations.sh -n $NO_OF_MIGRATIONS
# generate test migrtations in append test
# ./test/performance/generate-test-migrations.sh -n $NO_OF_MIGRATIONS -a

./migrator -configFile=test/performance/migrator-performance.yaml &> /dev/null &
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
start=`date +%s`
curl -d @create_version.txt http://localhost:8888/v2/service | jq -r '.data.createVersion.summary'
end=`date +%s`

echo "Done, checking if all migrations were applied correctly..."

output=$(psql -U postgres -h 127.0.0.1 -p 5432 -d migrator -tAq -c "select name from migrator.migrator_tenants")

IFS=$'\n'; tenants=($output); unset IFS;

for tenant in "${tenants[@]}"; do
    count=$(psql -U postgres -h 127.0.0.1 -p 5432 -d migrator -tAq -c "select count(distinct col) from $tenant.table_for_inserts")
    tables_count=$(psql -U postgres -h 127.0.0.1 -p 5432 -d migrator -tAq -c "select count(*) from information_schema.tables where table_schema = '$tenant' and table_name like 'table_%'")

    if [[ $count -ne $NO_OF_MIGRATIONS+$EXISTING_INSERTS+1 ]]; then
      echo "[migrations inserts] error for $tenant, got $count, expected: $((NO_OF_MIGRATIONS+$EXISTING_INSERTS+1))"
    fi
    if [[ $tables_count -ne $NO_OF_MIGRATIONS/10+$EXISTING_TABLES+1 ]]; then
      echo "[migrations tables] error for $tenant, got $tables_count, expected: $((NO_OF_MIGRATIONS/10+$EXISTING_TABLES+1))"
    fi

done

echo "Test took $((end-start)) seconds"

# remove generated migrations and test request and kill migrator
rm -rf test/performance/migrations
rm create_version.txt
sleep 5
killall migrator


# prepare for append test
# 1. comment out lines:
# 87
#
# 2. RUN TEST
#
# 3. comment out lines:
# 10, 11
# 20, 22, 24
# 27
#
# 4. uncomment lines:
# 13, 14
# 29
#
# 5. RUN TEST
