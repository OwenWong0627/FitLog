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
