package models

type User struct {
	Id       int    `json:"id" validate:"-"`
	Role     int    `json:"role" validate:"required"`
	Name     string `json:"name" validate:"required,min=2,max=50"`
	Username string `json:"username" validate:"required,min=3,max=50"`
	Password string `json:"password" validate:"required,min=5,max=50"`
	Email    string `json:"email" validate:"required,email"`
}
