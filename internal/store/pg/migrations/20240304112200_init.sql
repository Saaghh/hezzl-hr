-- +migrate Up

CREATE TABLE projects (
    id bigserial not null unique primary key,
    name varchar not null,
    created_at timestamp with time zone default now()
);

CREATE TABLE goods (
    PRIMARY KEY (id, project_id),
    id bigserial not null,
    project_id bigserial not null references projects(id),
    name varchar not null,
    description varchar not null default '',
    priority int not null,
    removed boolean not null default false,
    created_at timestamp with time zone default now()
);

-- +migrate Down

DROP TABLE projects, goods CASCADE;