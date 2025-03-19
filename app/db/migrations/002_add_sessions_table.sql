-- +goose Up
create table if not exists sessions(
	id integer primary key,
	token text not null,
	user_id integer not null references users(id),
	ip_address text,
	user_agent text,
	expires_at datetime not null, 
	created_at datetime not null, 
    updated_at datetime not null, 
	deleted_at datetime 
);
CREATE INDEX idx_sessions_token ON sessions(token);
CREATE INDEX idx_sessions_user_id ON sessions(user_id);

-- +goose Down
drop table if exists sessions;
