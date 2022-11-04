package models

import (
	"github.com/fernandotsda/nemesys/shared/types"
)

type Metric[T any] struct {
	// Base is the base metric configuration.
	Base BaseMetric `json:"base" validate:"required"`
	// Protocol is the protocol configuration.
	Protocol T `json:"protocol" validate:"required"`
}

type BaseMetricSimplified struct {
	// Id is the metric unique identifier.
	Id int64 `json:"id" validate:"-"`
	// ContainerId is the metric container identifier.
	ContainerId int32 `json:"container-id" validate:"required"`
	// ContainerType is the metric container type.
	ContainerType types.ContainerType `json:"container-type" validate:"-"`
	// Name is the metric name.
	Name string `json:"name" validate:"required,max=50"`
	// Ident is the metric string identification.
	Ident string `json:"ident" validate:"required,max=50"`
	// Descr is the metric description.
	Descr string `json:"descr" validate:"required,max=255"`
}

type BaseMetric struct {
	// Id is the metric unique identifier.
	Id int64 `json:"id" validate:"-"`
	// ContainerId is the metric container identifier.
	ContainerId int32 `json:"container-id" validate:"-"`
	// ContainerType is the metric container type.
	ContainerType types.ContainerType `json:"container-type" validate:"-"`
	// Type is the metric type.
	Type types.MetricType `json:"type" validate:"required"`
	// Name is the metric name.
	Name string `json:"name" validate:"required,max=50"`
	// Ident is the metric string identification.
	Ident string `json:"ident" validate:"required,max=50"`
	// Descr is the metric description.
	Descr string `json:"descr" validate:"required,max=255"`
	// DataPolicyId is the metric data policy identifier.
	DataPolicyId int16 `json:"data-policy-id" validate:"required"`
	// RTSPullingInterval is the interval in miliseconds between each pull. Max is one hour.
	RTSPullingInterval int32 `json:"rts-pulling-interval" validate:"required,max=3600000"`
	// RTSPullingTimes is how many times will pull the data.
	RTSPullingTimes int16 `json:"rts-pulling-times" validate:"max=1000000"`
	// RTSCacheDuration is the data duration in miliseconds on RTS cache. Max is one hour.
	RTSCacheDuration int32 `json:"rts-cache-duration" validate:"max=3600000"`
	// EvaluableExpression is the a evaluable expression for the metric value.
	EvaluableExpression string `json:"evaluable-expression" validate:"max=255"`
}

type MetricDataResponse struct {
	// ContainerId is the metric's container identifier.
	ContainerId int32
	// MetricId is the metric identifier.
	MetricId int64
	// MetricType is the data type.
	MetricType types.MetricType
	// Value is the metric data as bytes.
	Value any
}

type MetricRequest struct {
	// ContainerId is the metric's container identifier.
	ContainerId int32
	// ContainerType is the metric's container type.
	ContainerType types.ContainerType
	// MetricId is the metric identifier.
	MetricId int64
	// MetricType is the metric type.
	MetricType types.MetricType
}
