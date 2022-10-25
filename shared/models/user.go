package models

type User struct {
	// Id is the user identifier.
	Id int `json:"id" validate:"-"`
	// Role is the user role.
	Role uint8 `json:"role" validate:"required"`
	// Name is the user name.
	Name string `json:"name" validate:"required,min=2,max=50"`
	// Username is the user's username.
	Username string `json:"username" validate:"required,min=3,max=50"`
	// Password is the user's password.
	Password string `json:"password" validate:"required,min=5,max=50"`
	// Email is the user's email.
	Email string `json:"email" validate:"required,email"`
}

type UserSimplified struct {
	// Id is the user identifier.
	Id int `json:"id" validate:"-"`
	// Username is the user's username.
	Username string `json:"username" validate:"required,min=3,max=50"`
	// Name is the user name.
	Name string `json:"name" validate:"required,min=2,max=50"`
}

type UserWithoutPW struct {
	// Id is the user identifier.
	Id int `json:"id" validate:"-"`
	// Role is the user role.
	Role int `json:"role" validate:"required"`
	// Name is the user name.
	Name string `json:"name" validate:"required,min=2,max=50"`
	// Username is the user's username.
	Username string `json:"username" validate:"required,min=3,max=50"`
	// Email is the user's email.
	Email string `json:"email" validate:"required,email"`
}
