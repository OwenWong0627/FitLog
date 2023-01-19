package db

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

type OTP struct {
	Username string `json:"username"`
	OTP      string `json:"otp"`
}
