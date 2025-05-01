-- +goose up

ALTER TABLE settings ADD COLUMN business_name text;

-- +goose down
ALTER TABLE settings DROP COLUMN business_name;
