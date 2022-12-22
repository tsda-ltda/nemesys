package models

type Login struct {
	// Username is the username.
	Username string `json:"username" validate:"required,min=2,max=50"`
	// Password is the password.
	Password string `json:"password" validate:"required,min=5,max=50"`
}
