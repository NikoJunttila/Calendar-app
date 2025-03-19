-- +goose Up
create table if not exists work_resources(
	id integer primary key,
    name text not null,
	owner_id integer not null,
    calendar_id integer not null,
    resources_percentage integer not null,
	created_at datetime not null,
	updated_at datetime not null,
	deleted_at datetime,
	FOREIGN KEY (owner_id) REFERENCES users(id)
	FOREIGN KEY (calendar_id) REFERENCES calendars(id)
);

-- +goose Down
drop table if exists work_resources;
