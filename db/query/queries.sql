-- name: GetEntriesByTraceID :many
select * from entries
where trace_id = $1
order by created_at asc;

-- name: SearchLogs :many
select * from entries
where type = 'log'
  and (sqlc.narg('service')::text is null or service = sqlc.narg('service'))
  and (sqlc.narg('level')::text is null or level = sqlc.narg('level'))
  and (sqlc.narg('query')::text is null or message ilike '%' || sqlc.narg('query') || '%')
  and (sqlc.narg('from')::timestamptz is null or created_at >= sqlc.narg('from'))
  and (sqlc.narg('to')::timestamptz is null or created_at < sqlc.narg('to'))
  and (sqlc.narg('before_id')::bigint is null or id < sqlc.narg('before_id'))
order by created_at desc, id desc
limit sqlc.arg('limit');
