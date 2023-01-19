package api

import (
	"crypto/hmac"
	"crypto/rsa"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"goapp/packages/config"
	"goapp/packages/db"
	"io/ioutil"
	"math/big"
	"net/http"

	"github.com/apex/log"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	cognito "github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"github.com/golang-jwt/jwt"
	_ "github.com/lib/pq"
)

type MyUser db.User
type MyResponse db.Response

type App struct {
	CognitoClient   *cognito.CognitoIdentityProvider
	CognitoRegion   string
	UserPoolID      string
	AppClientID     string
	AppClientSecret string
	jwk             *JWK
	jwkURL          string
	Token           string
}

type KeySet struct {
	Alg string `json:"alg"`
	E   string `json:"e"`
	Kid string `json:"kid"`
	Kty string `json:"kty"`
	N   string `json:"n"`
}

type JWK struct {
	Keys []KeySet `json:"keys"`
}

var server *fiber.App

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
			Region:      aws.String("us-east-1"),
			Credentials: credentials.NewStaticCredentials(config.Config[config.AWS_ACCESS_KEY], config.Config[config.AWS_SECRET_KEY], ""),
		},
		SharedConfigState: session.SharedConfigEnable,
	}))

	userPoolID := config.Config[config.USER_POOLID]
	appClientID := config.Config[config.APP_CLIENTID]
	appClientSecret := config.Config[config.APP_CLIENTSECRET]

	// Fill App structure with environment keys and session generated
	a := App{
		CognitoClient:   cognito.New(sess),
		CognitoRegion:   "us-east-1",
		UserPoolID:      userPoolID,
		AppClientID:     appClientID,
		AppClientSecret: appClientSecret,
	}

	a.jwkURL = fmt.Sprintf("https://cognito-idp.%s.amazonaws.com/%s/.well-known/jwks.json", a.CognitoRegion, a.UserPoolID)
	jwkerr := a.CacheJWK()
	if jwkerr != nil {
		log.WithField("reason", err.Error()).Fatal("caching jwk failed")
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

	api.Post("/login", WithDB(a.Login, conn))
	api.Post("/register", WithDB(a.CreateUser, conn))
	// api.Post("/otp", OTP)
	api.Get("/logout", a.Logout)

	// authed routes
	api.Get("/workouts", WithDB(a.Workouts, conn))
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

func (a *App) CreateUser(c *fiber.Ctx, dbConn *sql.DB) error {

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

	user := &cognito.SignUpInput{
		Username: aws.String(u.Username),
		Password: aws.String(u.Password),
		ClientId: aws.String(a.AppClientID),
		UserAttributes: []*cognito.AttributeType{
			{
				Name:  aws.String("email"),
				Value: aws.String(u.Email),
			},
		},
	}

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

func (a *App) Workouts(c *fiber.Ctx, dbConn *sql.DB) error {
	tokenString := c.Get("Authorization")
	if tokenString == "" {
		return c.SendStatus(http.StatusUnauthorized)
	}

	token, err := a.ParseJWT(tokenString)
	if err != nil {
		return c.Status(http.StatusBadRequest).SendString(err.Error())
	}

	if !token.Valid {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"success": false, "errors": []string{"Invalid Token"}})
	}

	claims := token.Claims.(jwt.MapClaims)
	userName, ok := claims["cognito:username"].(string)

	if !ok {
		return c.SendStatus(http.StatusUnauthorized)
	}
	user := &db.User{}
	if err := dbConn.QueryRow(db.GetUserByUsernameQuery, userName).
		Scan(&user.ID, &user.Username, &user.Email, &user.CreatedAt, &user.UpdatedAt); err != nil {
		if err == sql.ErrNoRows {
			return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"success": false, "errors": []string{"User DNE"}})
		}
	}

	return c.JSON(&fiber.Map{"success": true, "user": user})
}

func (a *App) Login(c *fiber.Ctx, dbConn *sql.DB) error {
	u := new(MyUser)

	if err := c.BodyParser(u); err != nil {
		return err
	}

	params := map[string]*string{
		"USERNAME": aws.String(u.Username),
		"PASSWORD": aws.String(u.Password),
	}

	secretHash := computeSecretHash(a.AppClientSecret, u.Username, a.AppClientID)
	params["SECRET_HASH"] = aws.String(secretHash)

	authTry := &cognito.InitiateAuthInput{
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
		return c.Status(http.StatusInternalServerError).JSON(authResp)
	}

	dbConn.Exec(db.UpdateLoginTime, u.Username)

	userFetch := &db.User{}
	if err := dbConn.QueryRow(db.GetUserByUsernameQuery, u.Username).
		Scan(&userFetch.ID, &userFetch.Username, &userFetch.Email, &userFetch.CreatedAt, &userFetch.UpdatedAt); err != nil {
		if err == sql.ErrNoRows {
			return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"success": false, "errors": []string{"Incorrect credentials"}})
		}
	}

	a.Token = *authResp.AuthenticationResult.IdToken

	return c.JSON(&fiber.Map{"success": true, "user": *userFetch, "token": a.Token, "authResp": authResp})
}

func (a *App) Logout(c *fiber.Ctx) error {
	// URI must change depending on which platform the app is hosted on
	logoutURI := "https://mydomain.auth.us-east-1.amazoncognito.com/logout?" + "client_id=" + a.AppClientID + "&logout_uri=" + config.CLIENT_URL + "login"
	fmt.Println(logoutURI)
	_, err := http.Get(logoutURI)
	if err != nil {
		return c.Status(http.StatusBadRequest).SendString(err.Error())
	}

	return c.JSON(&fiber.Map{"success": true})
}

func WithDB(fn func(c *fiber.Ctx, db *sql.DB) error, db *sql.DB) func(c *fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		return fn(c, db)
	}
}

// MapKeys indexes each KeySet against its KID
func (jwk *JWK) MapKeys() map[string]KeySet {
	keymap := make(map[string]KeySet)
	for _, keys := range jwk.Keys {
		keymap[keys.Kid] = keys
	}
	return keymap
}

func (a *App) CacheJWK() error {
	req, err := http.NewRequest("GET", a.jwkURL, nil)
	if err != nil {
		return err
	}

	req.Header.Add("Accept", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	jwk := new(JWK)
	err = json.Unmarshal(body, jwk)
	if err != nil {
		return err
	}

	a.jwk = jwk
	return nil
}

func (a *App) ParseJWT(tokenString string) (*jwt.Token, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		kid, ok := token.Header["kid"].(string)
		if !ok {
			return nil, fmt.Errorf("getting kid; not a string")
		}
		keymap := a.jwk.MapKeys()
		keyset, ok := keymap[kid]
		if !ok {
			return nil, fmt.Errorf("keyset not found for kid %s", kid)
		}
		key := convertKey(keyset.E, keyset.N)
		return key, nil
	})
	if err != nil {
		return token, fmt.Errorf("parsing jwt; %w", err)
	}

	return token, nil
}

func convertKey(rawE, rawN string) *rsa.PublicKey {
	decodedE, err := base64.RawURLEncoding.DecodeString(rawE)
	if err != nil {
		panic(err)
	}
	if len(decodedE) < 4 {
		ndata := make([]byte, 4)
		copy(ndata[4-len(decodedE):], decodedE)
		decodedE = ndata
	}
	pubKey := &rsa.PublicKey{
		N: &big.Int{},
		E: int(binary.BigEndian.Uint32(decodedE[:])),
	}
	decodedN, err := base64.RawURLEncoding.DecodeString(rawN)
	if err != nil {
		panic(err)
	}
	pubKey.N.SetBytes(decodedN)
	return pubKey
}
