create table if not exists migrator_tenants (
  id serial primary key,
  name varchar(200) not null,
  created timestamp default (now() at time zone 'utc')
);

insert into migrator_tenants (name) values ('abc');
insert into migrator_tenants (name) values ('def');
insert into migrator_tenants (name) values ('xyz');

create schema abc;
create schema def;
create schema xyz;
