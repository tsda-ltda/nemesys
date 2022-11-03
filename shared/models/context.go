package models

type Context struct {
	// Id is the identifier.
	Id int32 `json:"id" validate:"-"`
	// TeamId is the team identifier.
	TeamId int32 `json:"team-id" validate:"-"`
	// Name is the context name.
	Name string `json:"name" validate:"required,min=2,max=50"`
	// Ident is the context ident.
	Ident string `json:"ident" validate:"required,min=2,max=50"`
	// Descr is the context description.
	Descr string `json:"descr" validate:"max=255"`
}
