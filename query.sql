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

-- name: ExactMatch :one
select
    *
from
    links
where
    source = ?;

-- name: FuzzyMatch :many
select
    *
from
    links
where
    ? like '%' || replace (links.source, '%s', '%') || '%'
    and not source = ?;

-- name: CreatePageView :one
insert into
    views (path, ip)
values
    (?, ?) returning *;