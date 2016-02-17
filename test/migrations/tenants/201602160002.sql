create table {schema}.module (id integer, id_config integer);

alter table {schema}.module add foreign key (id_config) references public.config(id);
