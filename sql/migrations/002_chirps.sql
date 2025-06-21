-- +goose Up
create table chirps (
    id uuid primary key,
    created_at timestamp not null,
    updated_at timestamp not null,
    user_id uuid not null references users(id) on delete cascade,
    body text not null
);

-- +goose Down
drop table chirps;