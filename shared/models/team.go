package models

type Team struct {
	Id       int    `json:"id" validate:"-"`
	UsersIds []int  `json:"users-ids" validate:"required"`
	Name     string `json:"name" validate:"required,min=2,max=50"`
	Ident    string `json:"ident" validate:"required,min=2,max=50"`
	Descr    string `json:"descr" validate:"max=255"`
}
