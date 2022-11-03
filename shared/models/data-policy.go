package models

type DataPolicy struct {
	// Id is the identifier.
	Id int16 `json:"id" validate:"-"`
	// Descr is the data policy description.
	Descr string `json:"descr" validate:"required,min=2,max=255"`
	// UseAggregation is the aggregation enabled status.
	UseAggregation bool `json:"use-aggregation" validate:"-"`
	// Retention is the raw data retention in hours.
	Retention int32 `json:"retention" validate:"required,min=1"`
	// AggregationRetention is the aggregation retention in hours.
	AggregationRetention int32 `json:"aggregation-retention" validate:"required,min=1"`
	// AggregationInterval is the aggregation interval in seconds.
	AggregationInterval int32 `json:"aggregation-interval" validate:"required,min=1"`
}
