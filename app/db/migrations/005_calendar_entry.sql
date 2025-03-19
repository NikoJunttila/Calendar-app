-- +goose Up
create table if not exists calendar_entries(
	id integer primary key,
	calendar_id integer not null,
	work_resource_id integer not null,
	date datetime not null,
	year integer not null,
	month integer not null,
	week integer not null,
	hours float not null default 0,
	text text not null,
	created_at datetime not null,
	updated_at datetime not null,
	deleted_at datetime,
	FOREIGN KEY (calendar_id) REFERENCES calendars(id),
	FOREIGN KEY (work_resource_id) REFERENCES work_resources(id)
);

-- +goose Down
drop table if exists calendar_entries;