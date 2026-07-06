-- Tables
CREATE TABLE IF NOT EXISTS clients (
    id SERIAL PRIMARY KEY,
    phone VARCHAR(20) UNIQUE NOT NULL,
    name VARCHAR(100),
    client_type VARCHAR(20) DEFAULT 'novice',
    role VARCHAR(20) DEFAULT 'client',
    instructor_id INTEGER REFERENCES instructors(id),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS zones (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    max_capacity INTEGER NOT NULL,
    duration_minutes INTEGER NOT NULL
);

CREATE TABLE IF NOT EXISTS instructors (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    photo_url VARCHAR(500),
    rating DECIMAL(2,1)
);

CREATE TABLE IF NOT EXISTS equipment (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    type VARCHAR(50) NOT NULL,
    available_count INTEGER DEFAULT 0,
    price_per_slot DECIMAL(10,2) DEFAULT 0
);

CREATE TABLE IF NOT EXISTS slots (
    id SERIAL PRIMARY KEY,
    zone_id INTEGER REFERENCES zones(id),
    instructor_id INTEGER REFERENCES instructors(id),
    start_time TIMESTAMP NOT NULL,
    end_time TIMESTAMP NOT NULL,
    total_places INTEGER NOT NULL,
    free_places INTEGER NOT NULL,
    price DECIMAL(10,2) NOT NULL,
    status VARCHAR(20) DEFAULT 'available'
);

CREATE TABLE IF NOT EXISTS bookings (
    id SERIAL PRIMARY KEY,
    client_id INTEGER REFERENCES clients(id),
    slot_id INTEGER REFERENCES slots(id),
    equipment_type VARCHAR(10) DEFAULT 'own',
    equipment_id INTEGER REFERENCES equipment(id),
    price DECIMAL(10,2) NOT NULL,
    status VARCHAR(30) DEFAULT 'confirmed',
    cancellation_type VARCHAR(20),
    cancellation_reason TEXT,
    confirmation_deadline TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Seed data
INSERT INTO zones (name, description, max_capacity, duration_minutes) VALUES
    ('Болдеринг', 'Болдеринг - скалолазание на невысоких стенах без страховки', 8, 90),
    ('Трассы с верёвкой', 'Трассы с верхней страховкой для начинающих и опытных', 16, 120)
ON CONFLICT DO NOTHING;

INSERT INTO instructors (name, photo_url, rating) VALUES
    ('Александр Иванов', 'https://example.com/photos/alex.jpg', 4.8),
    ('Мария Петрова', 'https://example.com/photos/maria.jpg', 4.9),
    ('Дмитрий Сидоров', 'https://example.com/photos/dmitry.jpg', 4.7)
ON CONFLICT DO NOTHING;

INSERT INTO equipment (name, type, available_count, price_per_slot) VALUES
    ('Скальники', 'shoes', 10, 200),
    ('Страховочная система', 'harness', 8, 150),
    ('Обвязка', 'rope', 5, 300)
ON CONFLICT DO NOTHING;

-- Generate slots for next 7 days (corrected syntax)
-- Using DO block below instead

-- Fix: use explicit values for instructor_id
DO $$
DECLARE
    z_id INTEGER;
    i_id INTEGER;
    day_offset INTEGER;
    hour_offset INTEGER;
    z_duration INTEGER;
    z_max_capacity INTEGER;
    z_price DECIMAL(10,2);
    slot_start TIMESTAMP;
    slot_end TIMESTAMP;
BEGIN
    FOR z_id IN SELECT id FROM zones LOOP
        SELECT duration_minutes, max_capacity INTO z_duration, z_max_capacity FROM zones WHERE id = z_id;
        z_price := CASE WHEN (SELECT name FROM zones WHERE id = z_id) = 'Болдеринг' THEN 1000.00 ELSE 1500.00 END;
        
        FOR i_id IN SELECT id FROM instructors LOOP
            FOR day_offset IN 0..6 LOOP
                FOR hour_offset IN 0..5 LOOP
                    IF hour_offset * 2 < 12 THEN
                        slot_start := CURRENT_DATE + (day_offset || ' days')::INTERVAL + ((hour_offset * 2) || ' hours')::INTERVAL;
                        slot_end := slot_start + (z_duration || ' minutes')::INTERVAL;
                        
                        INSERT INTO slots (zone_id, instructor_id, start_time, end_time, total_places, free_places, price, status)
                        VALUES (z_id, i_id, slot_start, slot_end, z_max_capacity, z_max_capacity, z_price, 'available')
                        ON CONFLICT DO NOTHING;
                    END IF;
                END LOOP;
            END LOOP;
        END LOOP;
    END LOOP;
END $$;

-- Create default admin and trainer accounts
INSERT INTO clients (phone, name, role) VALUES
    ('+79999999999', 'Администратор', 'admin'),
    ('+79999999990', 'Тренер 1', 'trainer')
ON CONFLICT DO NOTHING;