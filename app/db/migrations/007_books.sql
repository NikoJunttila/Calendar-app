-- +goose Up
create table if not exists books (
    id integer primary key,
    name text not null,
    author text,
    owner_id integer not null,
    location text,
    pages integer,
	created_at datetime not null,
	updated_at datetime not null,
	deleted_at datetime,
	FOREIGN KEY (owner_id) REFERENCES users(id)
);
create table if not exists book_progress (
    id integer primary key,
    book_id integer not null,
    owner_id integer not null,
    current_page integer not null,
    last_read datetime not null,
    notes text,
    created_at datetime not null,
    updated_at datetime not null,
    FOREIGN KEY (book_id) REFERENCES books(id)
    FOREIGN KEY (owner_id) REFERENCES users(id)
);
-- +goose Down
drop table if exists books
drop table if exists book_progress