create table {schema}.config (id integer, key varchar(100), value varchar(100));

alter table {schema}.config add constraint pk_id primary key (id);
