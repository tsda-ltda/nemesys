package models

import "github.com/fernandotsda/nemesys/shared/types"

type Container[T any] struct {
	// Base is the base container settings.
	Base BaseContainer `json:"base" validate:"required"`
	// Protocol is the container protocol settings.
	Protocol T `json:"protocol" validate:"required"`
}

type BaseContainer struct {
	// Id is the unique identifier.
	Id int32 `json:"id" validate:"-"`
	// Name is the container name.
	Name string `json:"name" validate:"required,max=50"`
	// Ident is the cotnainer string identifier.
	Ident string `json:"ident" validate:"required,max=50"`
	// Descr is the container description.
	Descr string `json:"descr" validate:"required,max=255"`
	// Type is the container type.
	Type types.ContainerType `json:"type" validate:"-"`
}
