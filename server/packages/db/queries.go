package db

const (
	CheckUserExists         = `SELECT true from users WHERE email = $1`
	LoginQuery              = `SELECT * from users WHERE email = $1`
	DeleteUser              = `DELETE FROM users WHERE email = $1`
	CreateUserQuery         =  `INSERT INTO users(id, username, email) VALUES (DEFAULT, $1 , $2);`
	GetUserByIDQuery        =  `SELECT * FROM users WHERE id = $1`
	GetUserByEmailQuery     =  `SELECT * FROM users WHERE email = $1`
	GetUserByUsernameQuery  =  `SELECT * FROM users WHERE username = lower($1)`
)
