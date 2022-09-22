package models

type User struct {
	Id       int    `json:"id"`
	Role     int    `json:"role"`
	TeamsIds []int  `json:"teamsIds"`
	Name     string `json:"name"`
	Username string `json:"username"`
	Password string `json:"password"`
	Email    string `json:"email"`
}
