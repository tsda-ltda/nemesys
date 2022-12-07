package models

type AlarmProfile struct {
	// Id is the alarm profile unique identifier.
	Id int32 `json:"id" validate:"-"`
	// Name is the alarm profile name.
	Name string `json:"name" validate:"required,max=50"`
	// Descr is the alarm profile description.
	Descr string `json:"descr" validate:"required,max=255"`
}

type AlarmCategory struct {
	// Id is the alarm category unique identifier.
	Id int32 `json:"id" validate:"-"`
	// Name is the alarm category name.
	Name string `json:"name" validate:"required,max=50"`
	// Descr is the alarm category description.
	Descr string `json:"descr" validate:"required,max=255"`
	// Lever is the alarm level.
	Level int32 `json:"level" validate:"-"`
}

type AlarmExpression struct {
	// Id is the alarm expression unique identifier.
	Id int32 `json:"id" validate:"-"`
	// Name is the alarm expression name.
	Name string `json:"name" validate:"required,max=50"`
	// Expression is the alarm expression.
	Expression string `json:"expression" validate:"required,max=255"`
	// AlarmCategoryId is the alarm category id.
	AlarmCategoryId int32 `json:"alarm-category-id" validate:"required"`
}
