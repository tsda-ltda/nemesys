package models

type DataPolicy struct {
	// Id is the identifier.
	Id int16 `json:"id" validate:"-"`
	// Name is the data policy name.
	Name string `json:"name" validate:"required,max=50"`
	// Descr is the data policy description.
	Descr string `json:"descr" validate:"required,max=255"`
	// UseAggr is the aggregation enabled status.
	UseAggr bool `json:"use-aggregation" validate:"-"`
	// Retention is the raw data retention in hours.
	Retention int32 `json:"retention" validate:"required,min=1"`
	// AggrRetention is the aggregation retention in hours.
	AggrRetention int32 `json:"aggregation-retention" validate:"required,min=1"`
	// AggrInterval is the aggregation interval in seconds.
	AggrInterval int32 `json:"aggregation-interval" validate:"required,min=1"`
	// AggrFn is the aggregation funcion.
	AggrFn string `json:"aggregation-function" validate:"required"`
}
