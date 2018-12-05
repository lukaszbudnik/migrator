create schema [abc]
GO
create schema [def]
GO
create schema [xyz]
GO
create schema [migrator]
GO

IF NOT EXISTS (select * from information_schema.tables where table_schema = 'migrator' and table_name = 'migrator_tenants')
BEGIN
  create table [migrator].migrator_tenants (
    id int identity (1,1) primary key,
    name varchar(200) not null,
    created datetime default CURRENT_TIMESTAMP
  );

  insert into [migrator].migrator_tenants (name) values ('abc');
  insert into [migrator].migrator_tenants (name) values ('def');
  insert into [migrator].migrator_tenants (name) values ('xyz');

END

GO
