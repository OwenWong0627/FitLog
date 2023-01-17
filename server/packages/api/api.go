package api

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"github.com/apex/log"
	"goapp/packages/config"
	"goapp/packages/db"
	"database/sql"
	"net/http"
	"time"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	_ "github.com/lib/pq"
	"github.com/golang-jwt/jwt"
)
type MyApp db.App
type MyUser db.User
type MyResponse db.Response

type Claims struct {
	db.User
	jwt.StandardClaims
}

var server *fiber.App
var svc *cognitoidentityprovider.CognitoIdentityProvider

func WithDB(fn func(c *fiber.Ctx, db *sql.DB) error, db *sql.DB) func(c *fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		return fn(c, db)
	}
}


func StartServer() {
	// Connect to the PostgreSQL database
	conn, err := db.ConnectDB()
	if err != nil {
		log.WithField("reason", err.Error()).Fatal("Db connection error occurred")
	}
	defer conn.Close()

	// Migration
	runMigration := config.Config[config.RUN_MIGRATION]
	dbName := config.Config[config.POSTGRES_DB]
	port := config.Config[config.SERVER_PORT]

	if runMigration == "true" && conn != nil {
		if err := db.Migrate(conn, dbName); err != nil {
			log.WithField("reason", err.Error()).Fatal("db migration failed")
		}
	}
	
	// Create a new session with the AWS SDK and instantiate a new Cognito client
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		Config: aws.Config{
			Region: aws.String("us-east-1"),
			Credentials: credentials.NewStaticCredentials(config.Config[config.AWS_ACCESS_KEY], config.Config[config.AWS_SECRET_KEY], ""),
		},
		SharedConfigState: session.SharedConfigEnable,
	}))

	svc = cognitoidentityprovider.New(sess)

	userPoolID := config.Config[config.USER_POOLID]
	appClientID := config.Config[config.APP_CLIENTID]
	appClientSecret := config.Config[config.APP_CLIENTSECRET]

	// Fill App structure with environment keys and session generated
	a := MyApp{
		CognitoClient:   svc,
		UserPoolID:      userPoolID,
		AppClientID:     appClientID,
		AppClientSecret: appClientSecret,
	}

	// Set up HTTP server with fiber
	app := fiber.New()
	app.Use(logger.New())
	app.Use(requestid.New())

	// Set up CORS Middleware management
	api := app.Group("/api")
	api.Use(cors.New(cors.Config{
		AllowOrigins:     config.Config[config.CLIENT_URL],
		AllowCredentials: true,
		AllowHeaders:     "Content-Type, Content-Length, Accept-Encoding, Authorization, accept, origin",
		AllowMethods:     "POST, OPTIONS, GET, PUT",
		ExposeHeaders:    "Set-Cookie",
	}))

	// public
	api.Get("/ping", Pong)

	api.Post("/login", WithDB(a.Login, conn))
	api.Post("/register", WithDB(a.CreateUser, conn))
	// api.Post("/otp", OTP)
	api.Get("/logout", a.Logout)

	// authed routes
	api.Get("/workouts", AuthorizeSession, WithDB(Workouts, conn))
	server = app
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

func computeSecretHash(clientSecret string, username string, clientId string) string {
	mac := hmac.New(sha256.New, []byte(clientSecret))
	mac.Write([]byte(username + clientId))

	return base64.StdEncoding.EncodeToString(mac.Sum(nil))
}

func Pong(c *fiber.Ctx) error {
	return c.SendString("pong")
}

func (a *MyApp) CreateUser(c *fiber.Ctx, dbConn *sql.DB) error {

	r := new(MyResponse)
	u := new(MyUser)

	// Bind the user input saved in context to the u(User) variable and validate it
	if err := c.BodyParser(u); err != nil {
		return c.Status(http.StatusBadRequest).SendString(err.Error())
	}
	
	// Check for Duplicate usernames
	if err := dbConn.QueryRow(db.GetUserByUsernameQuery, u.Username).Scan(&u); err != nil {
		if err != sql.ErrNoRows {
			return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"success": false, "errors": []string{"Duplicate Usernames"}})
		}
	}

	// Check for Duplicate emails
	if err := dbConn.QueryRow(db.GetUserByEmailQuery, u.Email).Scan(&u); err != nil {
		if err != sql.ErrNoRows {
			return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"success": false, "errors": []string{"Duplicate Emails"}})
		}
	}

	user := &cognitoidentityprovider.SignUpInput{
		Username: aws.String(u.Username),
		Password: aws.String(u.Password),
		ClientId: aws.String(a.AppClientID),
		UserAttributes: []*cognitoidentityprovider.AttributeType{
			{
				Name:  aws.String("email"),
				Value: aws.String(u.Email),
			},
		},
	}
	fmt.Println(user)

	secretHash := computeSecretHash(a.AppClientSecret, u.Username, a.AppClientID)
	user.SecretHash = aws.String(secretHash)

	// Make signup operation using cognito's api
	_, r.Error = a.CognitoClient.SignUp(user)
	if r.Error != nil {
		return c.Status(http.StatusInternalServerError).JSON(r)
	}

	_, r.Error = dbConn.Query(db.CreateUserQuery, user.Username, user.UserAttributes[0].Value)
	if r.Error != nil {
			// handle error
			return c.Status(http.StatusUnauthorized).JSON(r)
	}

	return c.JSON(&fiber.Map{"success": true})
}



func Workouts(c *fiber.Ctx, dbConn *sql.DB) error {
	tokenUser := c.Locals("user").(*jwt.Token)
	claims := tokenUser.Claims.(jwt.MapClaims)
	fmt.Println("hi3")
	fmt.Println(claims)
	userName, ok := claims["username"].(string)

	if !ok {
		fmt.Println("hi")
		fmt.Println(claims)
		// fmt.Println(claims["email"].(string))
		return c.SendStatus(http.StatusUnauthorized)
	}
	fmt.Println("hi2")
	fmt.Println(claims)
	user := &db.User{}
	if err := dbConn.QueryRow(db.GetUserByUsernameQuery, userName).
		Scan(&user.ID, &user.Username, &user.Email, &user.CreatedAt, &user.UpdatedAt); err != nil {
		if err == sql.ErrNoRows {
			fmt.Println("hi1")
			fmt.Println(user)
			return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"success": false, "errors": []string{"User DNE"}})
		}
	}
	return c.JSON(&fiber.Map{"success": true, "user": user})
}



func (a *MyApp) Login(c *fiber.Ctx, dbConn *sql.DB) error {
	u := new(MyUser)
	// user := &db.User{}
	
	if err := c.BodyParser(u); err != nil {
		return err
	}

	fmt.Println(u)

	params := map[string]*string{
		"USERNAME": aws.String(u.Username),
		"PASSWORD": aws.String(u.Password),
	}

	secretHash := computeSecretHash(a.AppClientSecret, u.Username, a.AppClientID)
	params["SECRET_HASH"] = aws.String(secretHash)

	authTry := &cognitoidentityprovider.InitiateAuthInput{
		AuthFlow: aws.String("USER_PASSWORD_AUTH"),
		AuthParameters: map[string]*string{
			"USERNAME":    aws.String(*params["USERNAME"]),
			"PASSWORD":    aws.String(*params["PASSWORD"]),
			"SECRET_HASH": aws.String(*params["SECRET_HASH"]),
		},
		ClientId: aws.String(a.AppClientID),
	}

	authResp, err := a.CognitoClient.InitiateAuth(authTry)
	if err != nil {
		fmt.Println(err)
		return c.Status(http.StatusInternalServerError).JSON(authResp)
	}

	userFetch := &db.User{}
	fmt.Println(u.Username)
	if err := dbConn.QueryRow(db.GetUserByUsernameQuery, u.Username).
		Scan(&userFetch.ID, &userFetch.Username, &userFetch.Email, &userFetch.CreatedAt, &userFetch.UpdatedAt); err != nil {
		if err == sql.ErrNoRows {
			return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"success": false, "errors": []string{"Incorrect credentials"}})
		}
	}

	// proceed here if login is successful, on the AWS side
	// Set up cookies to set up log in duration
	//expiration time of the token ->30 mins
	expirationTime := time.Now().Add(30 * time.Minute)

	claims := &Claims{
		User: *userFetch,
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

	a.Token = *authResp.AuthenticationResult.AccessToken
	return c.JSON(&fiber.Map{"success": true, "user": claims.User, "token": tokenValue, "authResp": authResp})
}



func (a *MyApp) Logout(c *fiber.Ctx) error {
	logoutURI := "https://exampleusers.auth.na-east-1.amazoncognito.com/logout?" + "client_id=" + a.AppClientID + "&logout_uri=https://tiruma.io"
	_, err := http.Get(logoutURI)
	if err != nil {
		return c.Status(http.StatusBadRequest).SendString(err.Error())
	}

	return c.SendStatus(http.StatusOK)
}
