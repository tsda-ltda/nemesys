package influxdb

import "errors"

var (
	ErrUnsupportedMetricType       = errors.New("unsupported metric type")
	ErrTaskNotFound                = errors.New("task not foud")
	ErrMetricDataResponseIsFailed  = errors.New("metric data response is set as failed")
	ErrInvalidDuration             = errors.New("invalid duration")
	ErrInvalidQueryOptions         = errors.New("invalid query options")
	ErrInvalidRetentionRulesLength = errors.New("invalid retention rules length")
	ErrLogIsNil                    = errors.New("log is nil")
	ErrInvalidLog                  = errors.New("invalid log format")
	ErrInvalidCountReturn          = errors.New("fail to transform datapolicy points count into number")
)
