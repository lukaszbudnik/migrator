#!/bin/bash


while [[ "$#" > 0 ]]; do case $1 in
  -n|--number-of-migrations) no_of_migrations="$2"; shift;;
  -a|--append) append=1;;
  *) echo "Unknown parameter passed: $1"; exit 1;;
esac; shift; done

if [[ -z "$no_of_migrations" ]] || [[ -z "${no_of_migrations##*[!0-9]*}" ]]; then
  no_of_migrations=100
fi

if [[ -z "$append" ]]; then
  append=0
fi

function generate_first_table {
  year=$(date +'%Y')
  cat > "migrations/tenants/${year}0000000000.sql" <<EOL
create table :tenant.table_for_inserts (
col int
);
insert into :tenant.table_for_inserts values (-1);
EOL
}

function generate_table {
  i=$1
  counter=$2
  timestamp=$(date +'%Y%m%d%H%M%S%N')
  file="migrations/tenants/${timestamp}_${i}.sql"
  cat > $file <<EOL
create table :tenant.table_${counter} (
a int,
b float,
c varchar(100)
);
insert into :tenant.table_for_inserts values ($i);
EOL
}

function generate_alter_drop_inserts {
  i=$1
  counter=$2
  timestamp=$(date +'%Y%m%d%H%M%S%N')
  file="migrations/tenants/${timestamp}_${i}.sql"
  # if [[ $i%2 -eq 1 ]]; then
  #   echo "alter table :tenant.table_${counter} add column d int, add column e varchar, add column f int;" > $file
  # else
  #   echo "alter table :tenant.table_${counter} drop column d, drop column e, drop column f;" >> $file
  # fi
  # no_of_inserts=1000
  no_of_inserts=10
  while [[ $no_of_inserts -gt 0 ]]; do
    echo "insert into :tenant.table_${counter} (a, b, c) values ($RANDOM, $RANDOM, '$RANDOM');" >> $file
    let no_of_inserts-=1
  done
  echo "insert into :tenant.table_for_inserts values ($i);" >> $file
}

if [[ $append -eq 0 ]]; then
  rm -rf migrations/tenants
  mkdir -p migrations/tenants

  generate_first_table

  i=0
  counter=0
else
  i=$(ls -t migrations/tenants | head -1 | cut -d '_' -f 2 | cut -d '.' -f 1)
  counter=$((i/10+1))
  i=$((i+1))
fi

end=$((i+no_of_migrations))

echo "About to generate $no_of_migrations migrations"

while [[ $i -lt $end ]]; do
  if [[ $i%10 -eq 0 ]]; then
    let counter+=1
    echo "generate_table $i $counter"
    generate_table $i $counter
  else
    # echo "generate_alter_drop_inserts $i $counter"
    generate_alter_drop_inserts $i $counter
  fi
  let i+=1
done
