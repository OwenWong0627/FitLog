package db

import (
	"github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
)

type App struct {
	CognitoClient   *cognitoidentityprovider.CognitoIdentityProvider
	UserPoolID      string
	AppClientID     string
	AppClientSecret string
	Token           string
}

type User struct {
	ID        string `json:"id,omitempty"`
	Password  string `json:"password,omitempty"`
	Email     string `json:"email,omitempty"`
	Username  string `json:"username,omitempty"`
	CreatedAt string `json:"created_at,omitempty"`
	UpdatedAt string `json:"updated_at,omitempty"`
}

type ResetPassword struct {
	Username        string `json:"username"`
	Password        string `json:"password"`
	ConfirmPassword string `json:"confirm_password"`
}

type UserConfirmationCode struct {
	ConfirmationCode string `json:"confirmationCode" validate:"required"`
	User             User   `json:"user" validate:"required"`
}

type Response struct {
	Error error `json:"error"`
}
