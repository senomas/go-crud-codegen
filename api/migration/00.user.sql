CREATE TABLE IF NOT EXISTS app_user (
    id SERIAL PRIMARY KEY,
    email TEXT,
    name TEXT,
    salt TEXT,
    password TEXT,
    token TEXT
)
