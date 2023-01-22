CREATE TABLE IF NOT EXISTS workoutlogs (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id) NOT NULL,
    date TIMESTAMP NOT NULL default current_timestamp,
    exercise VARCHAR(255) NOT NULL,
    reps INTEGER NOT NULL,
    weightlbs DECIMAL(5, 2),
    weightkg DECIMAL(5, 2)
);

CREATE INDEX date_index ON workoutlogs (date);
