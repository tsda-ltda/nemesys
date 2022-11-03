package models

type Team struct {
	// Id is the identifier.
	Id int32 `json:"id" validate:"-"`
	// Name is the team name.
	Name string `json:"name" validate:"required,min=2,max=50"`
	// Ident is the team ident.
	Ident string `json:"ident" validate:"required,min=2,max=50"`
	// Descr is the team description.
	Descr string `json:"descr" validate:"max=255"`
}

type AddMemberReq struct {
	// UserId is the user identifier.
	UserId int32 `json:"user-id" validate:"required"`
}
