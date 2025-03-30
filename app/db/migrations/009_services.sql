-- +goose Up
-- Services table to define different types of appointments/services offered
CREATE TABLE services (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    name TEXT NOT NULL,
    description TEXT,
    duration INTEGER NOT NULL, -- Duration in minutes
    price DECIMAL(10,2),
    color TEXT, -- For calendar display/visual identification
    is_active BOOLEAN NOT NULL DEFAULT 1,
    buffer_before INTEGER DEFAULT 0, -- Buffer time before appointment (in minutes)
    buffer_after INTEGER DEFAULT 0,  -- Buffer time after appointment (in minutes)
    max_attendees INTEGER DEFAULT 1, -- For group appointments/services
    location TEXT, -- Where the service takes place
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- Create index for quick lookup by user
CREATE INDEX idx_services_user_id ON services(user_id);
CREATE INDEX idx_services_is_active ON services(is_active);

-- Update time_slots table to reference services
ALTER TABLE time_slots ADD COLUMN service_id INTEGER REFERENCES services(id);
CREATE INDEX idx_time_slots_service_id ON time_slots(service_id);

-- Also update bookings table to store which service was booked
ALTER TABLE bookings ADD COLUMN service_id INTEGER REFERENCES services(id);
CREATE INDEX idx_bookings_service_id ON bookings(service_id);

-- +goose Down
-- SQL in this section is executed when the migration is rolled back
DROP INDEX IF EXISTS idx_bookings_service_id;
DROP INDEX IF EXISTS idx_time_slots_service_id;
DROP INDEX IF EXISTS idx_services_is_active;
DROP INDEX IF EXISTS idx_services_user_id;

-- Remove the service_id column from bookings table
PRAGMA foreign_keys=off;
CREATE TABLE bookings_temp (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    time_slot_id INTEGER NOT NULL,
    client_name TEXT NOT NULL,
    client_email TEXT NOT NULL,
    client_phone TEXT,
    booking_ref TEXT NOT NULL UNIQUE,
    notes TEXT,
    status TEXT NOT NULL DEFAULT 'confirmed',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (time_slot_id) REFERENCES time_slots(id) ON DELETE CASCADE
);
INSERT INTO bookings_temp SELECT id, user_id, time_slot_id, client_name, client_email, client_phone, booking_ref, notes, status, created_at, updated_at FROM bookings;
DROP TABLE bookings;
ALTER TABLE bookings_temp RENAME TO bookings;
CREATE INDEX idx_bookings_user_id ON bookings(user_id);
CREATE INDEX idx_bookings_time_slot_id ON bookings(time_slot_id);
CREATE INDEX idx_bookings_booking_ref ON bookings(booking_ref);
CREATE INDEX idx_bookings_status ON bookings(status);
PRAGMA foreign_keys=on;

-- Remove the service_id column from time_slots table
PRAGMA foreign_keys=off;
CREATE TABLE time_slots_temp (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    date TEXT NOT NULL,
    time TEXT NOT NULL,
    duration INTEGER NOT NULL,
    is_booked BOOLEAN NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);
INSERT INTO time_slots_temp SELECT id, user_id, date, time, duration, is_booked, created_at, updated_at FROM time_slots;
DROP TABLE time_slots;
ALTER TABLE time_slots_temp RENAME TO time_slots;
CREATE INDEX idx_time_slots_user_id ON time_slots(user_id);
CREATE INDEX idx_time_slots_date ON time_slots(date);
CREATE INDEX idx_time_slots_is_booked ON time_slots(is_booked);
CREATE INDEX idx_time_slots_user_date ON time_slots(user_id, date);
PRAGMA foreign_keys=on;

-- Drop the services table
DROP TABLE IF EXISTS services;