package models

type DataPolicy struct {
	Id                   int    `json:"id" validate:"-"`
	Descr                string `json:"descr" validate:"required,min=2,max=255"`
	UseAggregation       bool   `json:"use-aggregation" validate:"-"`
	Retention            int    `json:"retention" validate:"required,min=1"`             // Raw data retantion in hours.
	AggregationRetention int    `json:"aggregation-retention" validate:"required,min=1"` // Aggregated data retation in hours.
	AggregationInterval  int    `json:"aggregation-interval" validate:"required,min=1"`  // Aggregated data interval in seconds.
}
