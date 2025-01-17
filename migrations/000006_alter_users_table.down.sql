ALTER TABLE IF EXISTS users
    -- ADD COLUMN salt VARCHAR(255),
    -- ADD COLUMN password VARCHAR(255),
    DROP COLUMN server_key,
    DROP COLUMN stored_key;