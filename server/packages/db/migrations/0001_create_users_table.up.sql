CREATE TABLE IF NOT EXISTS users (
    id serial PRIMARY KEY,
    username TEXT NOT NULL,
    email TEXT UNIQUE NOT NULL,
    created_on TIMESTAMP NOT NULL default current_timestamp,
    updated_at TIMESTAMP NOT NULL default current_timestamp
);
