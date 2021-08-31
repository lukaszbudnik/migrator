#!/bin/bash

cd "$(dirname "$0")"

remove_tenant_placeholder=0
append=0

while [[ "$#" > 0 ]]; do case $1 in
  -n|--number-of-migrations) no_of_migrations="$2"; shift;;
  -a|--append) append=1;;
  -f|--flyway) remove_tenant_placeholder=1;;
  *) echo "Unknown parameter passed: $1"; exit 1;;
esac; shift; done

if [[ -z "$no_of_migrations" ]] || [[ -z "${no_of_migrations##*[!0-9]*}" ]]; then
  no_of_migrations=100
fi

tenantplaceholder=":tenant"
tenantprefixplaceholder="$tenantplaceholder."
if (( remove_tenant_placeholder == 1 )); then
  tenantplaceholder=""
  tenantprefixplaceholder=""
fi

function generate_first_table {
  year=$(date +'%Y')

  if (( remove_tenant_placeholder == 0 )); then
    cat > "migrations/tenants/V${year}.1__0000000000.sql" <<EOL
create schema if not exists ${tenantplaceholder};
EOL
  fi

  cat >> "migrations/tenants/V${year}.1__0000000000.sql" <<EOL
create table ${tenantprefixplaceholder}table_for_inserts (
col int
);
insert into ${tenantprefixplaceholder}table_for_inserts values (-1);
EOL
}

function generate_table {
  i=$1
  counter=$2
  timestamp=$(date +'%Y%m%d%H%M%S%N')
  file="migrations/tenants/V${timestamp}.${counter}_${i}.sql"
  cat > $file <<EOL
create table ${tenantprefixplaceholder}table_${counter} (
a int,
b float,
c varchar(100)
);
insert into ${tenantprefixplaceholder}table_for_inserts values ($i);
EOL
}

function generate_public_table {
  timestamp=$(date +'%Y%m%d%H%M%S%N')
  file="migrations/public/V${timestamp}.sql"
  cat > $file <<EOL
create table if not exists public.table_for_inserts (
a int,
b float,
c varchar(100)
);
insert into public.table_for_inserts values (0);
EOL
}

function generate_alter_drop_inserts {
  i=$1
  counter=$2
  timestamp=$(date +'%Y%m%d%H%M%S%N')
  file="migrations/tenants/V${timestamp}.${counter}_${i}.sql"
  # if [[ $i%2 -eq 1 ]]; then
  #   echo "alter table ${tenantprefixplaceholder}table_${counter} add column d int, add column e varchar, add column f int;" > $file
  # else
  #   echo "alter table ${tenantprefixplaceholder}table_${counter} drop column d, drop column e, drop column f;" >> $file
  # fi
  # no_of_inserts=1000
  no_of_inserts=10
  while [[ $no_of_inserts -gt 0 ]]; do
    echo "insert into ${tenantprefixplaceholder}table_${counter} (a, b, c) values ($RANDOM, $RANDOM, '$RANDOM');" >> $file
    let no_of_inserts-=1
  done
  echo "insert into ${tenantprefixplaceholder}table_for_inserts values ($i);" >> $file
}

if [[ $append -eq 0 ]]; then
  rm -rf migrations
  mkdir -p migrations/tenants
  mkdir -p migrations/public

  generate_first_table
  generate_public_table

  i=0
  counter=0
else
  i=$(ls -t migrations/tenants | head -1 | cut -d '_' -f 2 | cut -d '.' -f 1)
  counter=$((i/10+1))
  i=$((i+1))
fi

end=$((i+no_of_migrations))

echo "About to generate $no_of_migrations migrations"
echo "is append? $append"
echo "is without tenant prefix? $remove_tenant_placeholder"
echo "counter = $counter"
echo "i = $i"

while [[ $i -lt $end ]]; do
  if [[ $i%10 -eq 0 ]]; then
    let counter+=1
    echo "generate_table $i $counter"
    generate_table $i $counter
  else
    #echo "generate_alter_drop_inserts $i $counter"
    generate_alter_drop_inserts $i $counter
  fi
  let i+=1
done
