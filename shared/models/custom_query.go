package models

type CustomQuery struct {
	// Id is the custom query unique id.
	Id int32 `json:"id" validate:"-"`
	// Ident is the custom query unique ident.
	Ident string `json:"ident" validate:"required,max=1000"`
	// Descr is the custom query description.
	Descr string `json:"descr" validate:"required"`
	// Flux is the flux code to use during the query.
	Flux string `json:"flux" validate:"required"`
}
