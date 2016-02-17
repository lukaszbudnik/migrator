alter table {schema}.users add column id_role integer;
alter table {schema}.users add foreign key (id_role) references ref.roles(id);
