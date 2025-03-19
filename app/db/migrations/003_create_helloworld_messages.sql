-- +goose Up
-- SQL in this section is executed when the migration is applied

-- Create helloworld_messages table
create table if not exists helloworld_messages(
    id integer primary key,
    message text not null,
    created_at datetime not null,
    updated_at datetime not null,
    deleted_at datetime
);

-- +goose Down
-- SQL in this section is executed when the migration is rolled back

drop table if exists helloworld_messages;