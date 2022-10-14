package models

type Context struct {
	Id     int    `json:"id" validate:"-"`
	TeamId int    `json:"team-id" validate:"-"`
	Name   string `json:"name" validate:"required,min=2,max=50"`
	Ident  string `json:"ident" validate:"required,min=2,max=50"`
	Descr  string `json:"descr" validate:"max=255"`
}

type ContextCreateReq struct {
	Name  string `json:"name" validate:"required,min=2,max=50"`
	Ident string `json:"ident" validate:"required,min=2,max=50"`
	Descr string `json:"descr" validate:"max=255"`
}
