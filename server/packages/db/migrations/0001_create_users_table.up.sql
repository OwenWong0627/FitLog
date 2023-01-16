CREATE TABLE IF NOT EXISTS workoutlogs (
    id SERIAL PRIMARY KEY,
    date TIMESTAMP NOT NULL,
    exercise VARCHAR(255) NOT NULL,
    sets INTEGER NOT NULL,
    reps INTEGER NOT NULL,
    weightlbs DECIMAL(5, 2) NOT NULL,
    weightkg DECIMAL(5, 2) NOT NULL
);

CREATE INDEX date_index ON workoutlogs (date);

CREATE TABLE IF NOT EXISTS lifetime_personal_record (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id) NOT NULL,
    exercise VARCHAR(255) NOT NULL,
    reps INTEGER NOT NULL,
    weightlbs DECIMAL(5, 2) NOT NULL,
    weightkg DECIMAL(5, 2) NOT NULL,
    created_on TIMESTAMP NOT NULL default current_timestamp,
    updated_at TIMESTAMP NOT NULL default current_timestamp
);

CREATE TABLE IF NOT EXISTS users (
    id serial PRIMARY KEY,
    username TEXT NOT NULL,
    email TEXT UNIQUE NOT NULL,
    created_on TIMESTAMP NOT NULL default current_timestamp,
    updated_at TIMESTAMP NOT NULL default current_timestamp
);