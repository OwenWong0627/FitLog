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
