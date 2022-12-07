package models

type Id32 struct {
	Id int32 `json:"id" validate:"required"`
}

type Id64 struct {
	Id int64 `json:"id" validate:"required"`
}
