package models

type User struct {
	// Id is the user identifier.
	Id int32 `json:"id" validate:"-"`
	// Role is the user role.
	Role uint8 `json:"role" validate:"required"`
	// FirstName is the user first name.
	FirstName string `json:"first-name" validate:"required,min=2,max=50"`
	// LastName is the user last name.
	LastName string `json:"last-name" validate:"required,min=2,max=50"`
	// Username is the user's username.
	Username string `json:"username" validate:"required,min=3,max=50"`
	// Password is the user's password.
	Password string `json:"password" validate:"required,min=5,max=50"`
	// Email is the user's email.
	Email string `json:"email" validate:"required,email,max=255"`
}

type UserWithoutPW struct {
	// Id is the user identifier.
	Id int32 `json:"id" validate:"-"`
	// Role is the user role.
	Role uint8 `json:"role" validate:"required"`
	// FirstName is the user first name.
	FirstName string `json:"first-name" validate:"required,min=2,max=50"`
	// LastName is the user last name.
	LastName string `json:"last-name" validate:"required,min=2,max=50"`
	// Username is the user's username.
	Username string `json:"username" validate:"required,min=3,max=50"`
	// Email is the user's email.
	Email string `json:"email" validate:"required,email"`
}
