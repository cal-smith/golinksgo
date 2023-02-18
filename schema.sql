PRAGMA journal_mode = WAL;

create table
    if not exists links (
        source text primary key not null,
        destination text not null,
        description text,
        created datetime default current_timestamp not null
    );

create table
    if not exists views (
        id integer primary key not null,
        path text not null,
        ip text not null,
        created datetime default current_timestamp not null
    );

create virtual table if not exists links_fts using fts5 (
    source,
    description,
    content = links,
    tokenize = "porter"
);

drop trigger if exists links_ai;

drop trigger if exists links_ad;

drop trigger if exists links_au;

create trigger links_ai after insert on links begin
insert into
    links_fts (rowid, source, description)
values
    (new.rowid, new.source, new.description);
end;

create trigger links_ad after delete on links begin
insert into
    links_fts (links_fts, rowid, source, description)
values
    ('delete', old.rowid, old.source, old.description);
end;

create trigger links_au after
update on links begin
insert into
    links_fts (links_fts, rowid, source, description)
values
    ('delete', old.rowid, old.source, old.description);

insert into
    links_fts (rowid, source, description)
values
    (new.rowid, new.source, new.description);
end;