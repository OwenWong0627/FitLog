package api

import (
	"github.com/apex/log"
	"github.com/gofiber/fiber/v2"
	"goapp/packages/config"
	"goapp/packages/db"
	"database/sql"
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
	"github.com/golang-jwt/jwt"
	_ "github.com/lib/pq"
)
// combine api.go and usecontroller.go

var server *fiber.App
var svc *cognitoidentityprovider.CognitoIdentityProvider

func StartServer() {
	// Connect to the PostgreSQL database
	conn, err := db.ConnectDB()
	if err != nil {
		log.WithField("reason", err.Error()).Fatal("Db connection error occurred")
	}
	defer conn.Close()

	// Create a new session with the AWS SDK and instantiate a new Cognito client
	sess := session.Must(session.NewSessionWithOptions(session.Options{
			SharedConfigState: session.SharedConfigEnable,
	}))

	svc = cognitoidentityprovider.New(sess)

	// Migration
	runMigration := config.Config[config.RUN_MIGRATION]
	dbName := config.Config[config.POSTGRES_DB]
	port := config.Config[config.SERVER_PORT]

	if runMigration == "true" && conn != nil {
		if err := db.Migrate(conn, dbName); err != nil {
			log.WithField("reason", err.Error()).Fatal("db migration failed")
		}
	}

	server = httpServer(conn)
	serverErr := server.Listen(port)
	if serverErr != nil {
		log.WithField("reason", serverErr.Error()).Fatal("Server error")
	}
}

func StopServer() {
	if server != nil {
		err := server.Shutdown()
		if err != nil {
			log.WithField("reason", err.Error()).Fatal("Shutdown server error")
		}
	}
}

type Claims struct {
	db.User
	jwt.StandardClaims
}

func Pong(c *fiber.Ctx) error {
	return c.SendString("pong")
}

func CreateUser(c *fiber.Ctx, dbConn *sql.DB) error {
	user := &db.User{}

	// Use the AWS Cognito client to create a new user in the user pool
	_, err := svc.SignUp(&cognitoidentityprovider.SignUpInput{
			ClientId:   aws.String("31it4g08bqhp7sm931g8ejti0m"),
			Password:   aws.String(user.Password),
			Username:   aws.String(user.Username),
			UserAttributes: []*cognitoidentityprovider.AttributeType{
					{
							Name:  aws.String("email"),
							Value: aws.String(user.Email),
					},
			},
	})
	if err != nil {
			// handle error
			return err
	}
	// Successful registration

	// Insert the user information into the PostgreSQL database
	_, err = dbConn.Query(db.CreateUserQuery, user.Username, user.Email)
	if err != nil {
			// handle error
			return err
	}
	// Successful insertion

	
	user := &db.User{}

	if err := c.BodyParser(user); err != nil {
		return err
	}

	if errs := utils.ValidateUser(*user); len(errs) > 0 {
		return c.Status(http.StatusUnprocessableEntity).JSON(&fiber.Map{"success": false, "errors": errs})
	}

	if user.UserExists(dbConn) {
		return c.Status(400).JSON(&fiber.Map{"success": false, "errors": []string{"email already exists"}})
	}

	user.HashPassword()
	_, err := dbConn.Query(db.CreateUserQuery, user.Name, user.Password, user.Email)
	if err != nil {
		return err
	}
	return c.JSON(&fiber.Map{"success": true})
}

func Session(c *fiber.Ctx, dbConn *sql.DB) error {
	tokenUser := c.Locals("user").(*jwt.Token)
	claims := tokenUser.Claims.(jwt.MapClaims)
	userID, ok := claims["id"].(string)

	if !ok {
		return c.SendStatus(http.StatusUnauthorized)
	}

	user := &db.User{}
	if err := dbConn.QueryRow(db.GetUserByIDQuery, userID).
		Scan(&user.ID, &user.Name, &user.Password, &user.Email, &user.CreatedAt, &user.UpdatedAt); err != nil {
		if err == sql.ErrNoRows {
			return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"success": false, "errors": []string{"Incorrect credentials"}})
		}
	}
	user.Password = ""
	return c.JSON(&fiber.Map{"success": true, "user": user})
}

func Login(c *fiber.Ctx, dbConn *sql.DB) error {
	loginUser := &db.User{}

	if err := c.BodyParser(loginUser); err != nil {
		return err
	}

	user := &db.User{}
	if err := dbConn.QueryRow(db.GetUserByEmailQuery, loginUser.Email).
		Scan(&user.ID, &user.Name, &user.Password, &user.Email, &user.CreatedAt, &user.UpdatedAt); err != nil {
		if err == sql.ErrNoRows {
			return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"success": false, "errors": []string{"Incorrect credentials"}})
		}
	}

	match := utils.ComparePassword(user.Password, loginUser.Password)
	if !match {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"success": false, "errors": []string{"Incorrect credentials"}})
	}

	//expiration time of the token ->30 mins
	expirationTime := time.Now().Add(30 * time.Minute)

	user.Password = ""
	claims := &Claims{
		User: *user,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	var jwtKey = []byte(config.Config[config.JWT_KEY])
	tokenValue, err := token.SignedString(jwtKey)

	if err != nil {
		return err
	}

	c.Cookie(&fiber.Cookie{
		Name:     "token",
		Value:    tokenValue,
		Expires:  expirationTime,
		Domain:   config.Config[config.CLIENT_URL],
		HTTPOnly: true,
	})

	return c.JSON(&fiber.Map{"success": true, "user": claims.User, "token": tokenValue})
}

func Logout(c *fiber.Ctx) error {
	c.ClearCookie()
	return c.SendStatus(http.StatusOK)
}
