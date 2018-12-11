create table {schema}.module (id integer, id_config integer, foreign key (id_config) references config.config(id));
