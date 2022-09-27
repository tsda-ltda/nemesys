package models

type ContextualMetric struct {
	Id          int    `json:"id" validate:"-"`
	ContextId   int    `json:"context-id" validate:"-"`
	ContainerId int    `json:"container-id" validate:"-"`
	MetricId    int    `json:"metric-id" validate:"-"`
	Ident       string `json:"ident" validate:"required,min=2,max=50"`
	Name        string `json:"name" validate:"required,min=2,max=50"`
	Descr       string `json:"descr" validate:"max=255"`
}
