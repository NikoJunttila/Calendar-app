-- +goose up

ALTER TABLE time_slots ADD COLUMN reminder_sent datetime;
ALTER TABLE time_slots ADD COLUMN real_time integer;

-- +goose down
ALTER TABLE time_slots DROP COLUMN reminder_sent;
ALTER TABLE time_slots DROP COLUMN real_time;
