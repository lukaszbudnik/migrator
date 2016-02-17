create schema if not exists public;

create table if not exists public.migrator_tenants (
  id serial primary key,
  name varchar(200) not null,
  created timestamp default now()
);

insert into public.migrator_tenants (name) values ('abc');
insert into public.migrator_tenants (name) values ('def');
insert into public.migrator_tenants (name) values ('xyz');

create schema abc;
create schema def;
create schema xyz;
