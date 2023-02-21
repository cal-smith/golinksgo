// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.17.0
// source: query.sql

package links

import (
	"context"
	"database/sql"
	"time"
)

const createPageView = `-- name: CreatePageView :one
insert into
    views (path, ip)
values
    (?, ?) returning id, path, ip, created
`

type CreatePageViewParams struct {
	Path string
	Ip   string
}

func (q *Queries) CreatePageView(ctx context.Context, arg CreatePageViewParams) (View, error) {
	row := q.db.QueryRowContext(ctx, createPageView, arg.Path, arg.Ip)
	var i View
	err := row.Scan(
		&i.ID,
		&i.Path,
		&i.Ip,
		&i.Created,
	)
	return i, err
}

const createlink = `-- name: Createlink :one
insert into
    links (source, destination, description)
values
    (?, ?, ?) returning source, destination, description, created
`

type CreatelinkParams struct {
	Source      string
	Destination string
	Description sql.NullString
}

func (q *Queries) Createlink(ctx context.Context, arg CreatelinkParams) (Link, error) {
	row := q.db.QueryRowContext(ctx, createlink, arg.Source, arg.Destination, arg.Description)
	var i Link
	err := row.Scan(
		&i.Source,
		&i.Destination,
		&i.Description,
		&i.Created,
	)
	return i, err
}

const deleteLink = `-- name: DeleteLink :exec
delete from 
    links
where 
    source = ?
`

func (q *Queries) DeleteLink(ctx context.Context, source string) error {
	_, err := q.db.ExecContext(ctx, deleteLink, source)
	return err
}

const deleteView = `-- name: DeleteView :exec
delete from
    views
where
    path = ?
`

func (q *Queries) DeleteView(ctx context.Context, path string) error {
	_, err := q.db.ExecContext(ctx, deleteView, path)
	return err
}

const exactMatch = `-- name: ExactMatch :one
select
    source, destination, description, created
from
    links
where
    source = ?
`

func (q *Queries) ExactMatch(ctx context.Context, source string) (Link, error) {
	row := q.db.QueryRowContext(ctx, exactMatch, source)
	var i Link
	err := row.Scan(
		&i.Source,
		&i.Destination,
		&i.Description,
		&i.Created,
	)
	return i, err
}

const fuzzyMatch = `-- name: FuzzyMatch :many
select
    source,
    destination
from
    links
where
    @source like '%' || replace (links.source, '%s', '%') || '%'
    and not source = @source
`

type FuzzyMatchParams struct {
	Source  string
	Column2 interface{}
}

type FuzzyMatchRow struct {
	Source      string
	Destination string
}

func (q *Queries) FuzzyMatch(ctx context.Context, arg FuzzyMatchParams) ([]FuzzyMatchRow, error) {
	rows, err := q.db.QueryContext(ctx, fuzzyMatch, arg.Source, arg.Column2)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []FuzzyMatchRow
	for rows.Next() {
		var i FuzzyMatchRow
		if err := rows.Scan(&i.Source, &i.Destination); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const listTop = `-- name: ListTop :many
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
    ?
`

type ListTopRow struct {
	Source      string
	Destination string
	Created     time.Time
	Description sql.NullString
	Total       interface{}
}

func (q *Queries) ListTop(ctx context.Context, limit int64) ([]ListTopRow, error) {
	rows, err := q.db.QueryContext(ctx, listTop, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []ListTopRow
	for rows.Next() {
		var i ListTopRow
		if err := rows.Scan(
			&i.Source,
			&i.Destination,
			&i.Created,
			&i.Description,
			&i.Total,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const listlinks = `-- name: Listlinks :many
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
    created desc
`

type ListlinksRow struct {
	Source      string
	Destination string
	Created     time.Time
	Description sql.NullString
	Total       interface{}
}

func (q *Queries) Listlinks(ctx context.Context) ([]ListlinksRow, error) {
	rows, err := q.db.QueryContext(ctx, listlinks)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []ListlinksRow
	for rows.Next() {
		var i ListlinksRow
		if err := rows.Scan(
			&i.Source,
			&i.Destination,
			&i.Created,
			&i.Description,
			&i.Total,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const updateLink = `-- name: UpdateLink :one
update
    links
set
    source = ?,
    destination = ?,
    description = ?
where
    source = ?
returning source, destination, description, created
`

type UpdateLinkParams struct {
	Source      string
	Destination string
	Description sql.NullString
	Source_2    string
}

func (q *Queries) UpdateLink(ctx context.Context, arg UpdateLinkParams) (Link, error) {
	row := q.db.QueryRowContext(ctx, updateLink,
		arg.Source,
		arg.Destination,
		arg.Description,
		arg.Source_2,
	)
	var i Link
	err := row.Scan(
		&i.Source,
		&i.Destination,
		&i.Description,
		&i.Created,
	)
	return i, err
}
