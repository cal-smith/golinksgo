-- name: Listlinks :many
select
    source,
    destination,
    created,
    description,
    ifnull (
        (
            select
                count(*)
            from
                views
            where
                path = links.source
            group by
                path
        ),
        0
    ) as total
from
    links
order by
    created desc;

-- name: ListTop :many
select
    source,
    destination,
    created,
    description,
    ifnull (
        (
            select
                count(*)
            from
                views
            where
                path = links.source
            group by
                path
        ),
        0
    ) as total
from
    links
order by
    total desc
limit
    ?;

-- name: Createlink :one
insert into
    links (source, destination, description)
values
    (?, ?, ?) returning *;

-- name: DeleteLink :exec
delete from 
    links
where 
    source = ?;

-- name: UpdateLink :one
update
    links
set
    source = ?,
    destination = ?,
    description = ?
where
    source = ?
returning *;

-- name: ExactMatch :one
select
    *
from
    links
where
    source = ?;

-- name: FuzzyMatch :many
select
    source,
    destination
from
    links
where
    @source like '%' || replace (links.source, '%s', '%') || '%'
    and not source = @source;

-- name: CreatePageView :one
insert into
    views (path, ip)
values
    (?, ?) returning *;

-- name: DeleteView :exec
delete from
    views
where
    path = ?;