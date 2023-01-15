package links

import (
	"context"
	"database/sql"
)

const search = `-- name: Search :many
with search_results as (
	select
		source,
		description,
		rank
	from
		links_fts(?)
	union
	select
		source,
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
		) as rank
	from
		links
	where
		source like "%" || ? || "%"
	order by
		rank desc
), top_results as (
	select
		source,
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
		) as rank
	from
		links
	order by
		rank desc
)
select * from search_results
union
select * from top_results
order by
	rank desc
limit ?;
`

type SearchRow struct {
	Source      string
	Description sql.NullString
	Rank        int64
}

func (q *Queries) Search(ctx context.Context, term string, limit int64) ([]SearchRow, error) {
	rows, err := q.db.QueryContext(ctx, search, term, term, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []SearchRow
	for rows.Next() {
		var i SearchRow
		if err := rows.Scan(
			&i.Source,
			&i.Description,
			&i.Rank,
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
