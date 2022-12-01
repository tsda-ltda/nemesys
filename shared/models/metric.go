package models

import (
	"github.com/fernandotsda/nemesys/shared/types"
)

type MetricPairId struct {
	// Id is the metric unique identifier.
	Id int64
	// ContainerId is the container unique identtifier.
	ContainerId int32
}

type BasicMetricAddDataForm struct {
	// MetricId is the metric id.
	MetricId int64
	// MetricType is the metric type.
	MetricType types.MetricType
	// ContainerId is the container id.
	ContainerId int32
	// DataPolicyId is the data policy id.
	DataPolicyId int16
	// Enabled is the metric enabled status.
	Enabled bool
	// DHSEnabled is the DHS enabled status.
	DHSEnabled bool
}

type MetricDataByRefkey struct {
	// Refkye is the metric reference key.
	Refkey string `json:"refkey" validate:"required,max=200"`
	// Value is the value.
	Value any `json:"value" validate:"-"`
	// Timestamp is the timestamp in UNIX Epoch format
	Timestamp int64 `json:"timestamp" validate:"min=0"`
}

type MetricRefkey struct {
	// Id is the metric refkey unique identifier.
	Id int64 `json:"id" validate:"-"`
	// Refkey is the reference key.
	Refkey string `json:"refkey" validate:"required,max=200"`
	// MetricId is the metric id.
	MetricId int64 `json:"metric-id" validate:"-"`
}

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
	// Descr is the metric description.
	Descr string `json:"descr" validate:"required,max=255"`
	// Enabled is the metric enable state.
	Enabled bool `json:"enabled" validade:"-"`
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
	// Descr is the metric description.
	Descr string `json:"descr" validate:"required,max=255"`
	// Enabled is the metric enable state.
	Enabled bool `json:"enabled" validade:"-"`
	// CheckAlarm is the metric check alarm state.
	CheckAlarm bool `json:"check-alarm" validate:"-"`
	// DataPolicyId is the metric data policy identifier.
	DataPolicyId int16 `json:"data-policy-id" validate:"required"`
	// RTSPullingTimes is how many times will pull the data.
	RTSPullingTimes int16 `json:"rts-pulling-times" validate:"min=0,max=1000000"`
	// RTSCacheDuration is the data duration in miliseconds on RTS cache. Max is one hour.
	RTSCacheDuration int32 `json:"rts-cache-duration" validate:"min=1000,max=3600000"`
	// DHSEnabled is the enabled state of for the data history service.
	DHSEnabled bool `json:"dhs-enabled" validate:"-"`
	// DHSInterval is the interval in seconds of the data pulling of the data history service.
	DHSInterval int32 `json:"dhs-interval" validate:"-"`
	// EvaluableExpression is the a evaluable expression for the metric value.
	EvaluableExpression string `json:"evaluable-expression" validate:"max=255"`
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
	// DataPolicyId is the data policy id.
	DataPolicyId int16
	// CheckAlarm is the check alarm state.
	CheckAlarm bool
}

type MetricsRequest struct {
	// ContainerId is the metric's container identifier.
	ContainerId int32
	// ContainerType is the metric's container type.
	ContainerType types.ContainerType
	// Metrics is the metrics.
	Metrics []MetricBasicRequestInfo
}

type MetricBasicRequestInfo struct {
	// Id is the metric identifier.
	Id int64
	// Type is the metric type.
	Type types.MetricType
	// DataPolicyId is the data policy id.
	DataPolicyId int16
	// CheckAlarm is the check alarm state.
	CheckAlarm bool
}

type MetricDataResponse struct {
	MetricBasicDataReponse
	// ContainerId is the metric's container identifier.
	ContainerId int32
}

type MetricsDataResponse struct {
	// ContainerId is the metric's container identifier.
	ContainerId int32
	// Metrics is the metrics responses.
	Metrics []MetricBasicDataReponse
}

type MetricBasicDataReponse struct {
	// Id is the metric identifier.
	Id int64
	// Type is the data type.
	Type types.MetricType
	// Value is the metric data as MetricType.
	Value any
	// DataPolicyId is the data policy id.
	DataPolicyId int16
	// Failed is the failed status.
	Failed bool
}

type MetricEvaluableExpression struct {
	// Id is the metric identifier.
	Id int64
	// Expression is the metric expression.
	Expression string
}
