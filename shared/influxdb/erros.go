package influxdb

import "errors"

var (
	ErrUnsupportedMetricType = errors.New("unsupported metric type")
	ErrTaskNotFound          = errors.New("task not foud")
)
