-- name: CreateChirp :one
insert into chirps (id, created_at, updated_at, body, user_id)
values (
    gen_random_uuid(),
    now(),
    now(),
    $1,
    $2
)
returning *;

-- name: ListChirps :many
select
    id,
    created_at,
    updated_at,
    body,
    user_id
from
    chirps
order by created_at asc;

-- name: ListChirpsForAuthor :many
select
    id,
    created_at,
    updated_at,
    body,
    user_id
from
    chirps
where
    user_id = $1
order by created_at asc;

-- name: GetChirp :one
select
    id,
    created_at,
    updated_at,
    body,
    user_id
from
    chirps
where
    id = $1;

-- name: DeleteChirp :exec
delete from chirps
where id = $1;