ALTER TABLE IF EXISTS users
    -- DROP COLUMN password,
    ADD COLUMN server_key VARCHAR(255),
    ADD COLUMN stored_key VARCHAR(255);