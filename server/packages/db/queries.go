package db

const (
	CreateUserQuery        = `INSERT INTO users(id, username, email) VALUES (DEFAULT, $1 , $2);`
	GetUserByIDQuery       = `SELECT * FROM users WHERE id = $1`
	GetUserByEmailQuery    = `SELECT * FROM users WHERE email = $1`
	GetUserByUsernameQuery = `SELECT * FROM users WHERE lower(username) = lower($1)`
	UpdateLoginTime        = `UPDATE users SET updated_at=NOW() WHERE lower(username) = lower($1);`
	AddWorkout             = `INSERT INTO workoutlogs(id, user_id, date, exercise, reps, weightlbs, weightkg) VALUES (DEFAULT, $1, DEFAULT, $2, $3, $4, $5)`
	GetWorkoutsByTime      = `SELECT * FROM workoutlogs WHERE user_id = $ ORDER BY date`
)
