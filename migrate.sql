begin transaction;
create temporary table links_backup (
    id
    source,
    destination,
    created
);
insert into links_backup select id, source, destination, created from links;
drop table links;
create table links (
    source text primary key not null,
    destination text not null,
    created datetime default current_timestamp not null
);
insert into links select source, destination, created from links_backup;
drop table links_backup;
commit;