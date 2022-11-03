package models

type ContextualMetric struct {
	// Id is the contextual identifier.
	Id int64 `json:"id" validate:"-"`
	// ContextId is the context identifier.
	ContextId int32 `json:"context-id" validate:"-"`
	// MetricId is the metric identfier.
	MetricId int64 `json:"metric-id" validate:"required"`
	// Ident is the contextual metric ident.
	Ident string `json:"ident" validate:"required,min=2,max=50"`
	// Name is the contextual metric name.
	Name string `json:"name" validate:"required,min=2,max=50"`
	// Descr is the conextual metric description.
	Descr string `json:"descr" validate:"max=255"`
}

type Data struct {
	Value any `json:"value"`
}
