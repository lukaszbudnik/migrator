#!/bin/bash

# stop on error
set -e

function cleanup {
    rm versions.txt || true
    rm create_version.txt || true
    rm tenants.txt || true
    rm create_tenant.txt || true
}

# use migrator service built from local branch
MIGRATOR_PORT=8282

# used in create* methods
COMMIT_SHA=$(git rev-list -1 HEAD)

# 1. Fetch migrator versions

echo "------------------------------------------------------------------------------"
echo "1. About to fetch migrator versions..."
cat <<EOF | tr -d "\n" > versions.txt
{
  "query": "
  query Versions {
    versions {
        id,
        name,
        created,
      }
  }",
  "operationName": "Versions"
}
EOF
versions=$(curl -s -d @versions.txt http://localhost:$MIGRATOR_PORT/v2/service | jq -r ".data.versions")

versions_count_before=$(echo $versions | jq length)
echo "Number of versions in migrator: $versions_count_before"

# 2. Fetch tenants

echo "------------------------------------------------------------------------------"
echo "2. About to fetch tenants..."
cat <<EOF | tr -d "\n" > tenants.txt
{
  "query": "
  query Tenants {
    tenants {
        name
      }
  }",
  "operationName": "Tenants"
}
EOF
tenants=$(curl -s -d @tenants.txt http://localhost:$MIGRATOR_PORT/v2/service | jq -r ".data.tenants")

tenants_count_before=$(echo $tenants | jq length)
echo "Number of tenants in migrator: $tenants_count_before"

# 3. Create new migrator version

echo "------------------------------------------------------------------------------"
echo "3. About to create new migrator version..."
VERSION_NAME="create-version-$COMMIT_SHA-$RANDOM"
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
      "versionName": "$VERSION_NAME"
    }
  }
}
EOF
version_create=$(curl -s -d @create_version.txt http://localhost:$MIGRATOR_PORT/v2/service | jq -r '.data.createVersion')
version_create_name=$(echo $version_create | jq -r '.version.name')

if [ "$version_create_name" != "$VERSION_NAME" ]; then
    >&2 echo "Version created with name '$version_create_name' but expected to be '$VERSION_NAME'"
    cleanup
    exit 1
fi
echo "New version successfully created"
echo $version_create | jq

# 4. Fetch migrator versions - will now contain version created above

echo "------------------------------------------------------------------------------"
echo "4. About to fetch migrator versions..."
versions=$(curl -s -d @versions.txt http://localhost:$MIGRATOR_PORT/v2/service | jq -r ".data.versions")
versions_count=$(echo $versions | jq length)

if [ "$versions_count" -le "$versions_count_before" ]; then
    >&2 echo "Versions count '$versions_count' should be greater than '$versions_count_before'"
    cleanup
    exit 1
fi

echo "Number of versions in migrator: $versions_count"

# 5. Create new tenant

echo "------------------------------------------------------------------------------"
echo "5. About to create new tenant..."
VERSION_NAME="create-tenant-$COMMIT_SHA-$RANDOM"
TENANT_NAME="newcustomer$RANDOM"
cat <<EOF | tr -d "\n" > create_tenant.txt
{
  "query": "
  mutation CreateTenant(\$input: TenantInput!) {
    createTenant(input: \$input) {
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
  "operationName": "CreateTenant",
  "variables": {
    "input": {
      "versionName": "$VERSION_NAME",
      "tenantName": "$TENANT_NAME"
    }
  }
}
EOF
tenant_create=$(curl -s -d @create_tenant.txt http://localhost:$MIGRATOR_PORT/v2/service | jq -r '.data.createTenant')
tenant_create_version_name=$(echo $tenant_create | jq -r '.version.name')

if [ "$tenant_create_version_name" != "$VERSION_NAME" ]; then
    >&2 echo "Version created with name '$tenant_create_version_name' but expected to be '$VERSION_NAME'"
    cleanup
    exit 1
fi
echo "New tenant successfully created"
echo $tenant_create | jq

# 6. Fetch tenants - will now contain tenant create above

echo "------------------------------------------------------------------------------"
echo "6. About to fetch tenants..."
tenants=$(curl -s -d @tenants.txt http://localhost:$MIGRATOR_PORT/v2/service | jq -r ".data.tenants")
tenants_count=$(echo $tenants | jq length)

if [ "$tenants_count" -le "$tenants_count_before" ]; then
    >&2 echo "Tenant count '$tenants_count' should be greater than '$tenants_count_before'"
    cleanup
    exit 1
fi

echo "Number of tenants: $tenants_count"

echo "------------------------------------------------------------------------------"

echo "All good!"

cleanup