-- +goose Up
create table if not exists calendars(
	id integer primary key,
    name text not null,
    index_number integer default 10,
	owner_id integer not null,
	work boolean default false,
	daily_work_hours float default 7.25,
	created_at datetime not null,
	updated_at datetime not null,
	deleted_at datetime,
	FOREIGN KEY (owner_id) REFERENCES users(id)
);

-- +goose Down
drop table if exists calendars;
