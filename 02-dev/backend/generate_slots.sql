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