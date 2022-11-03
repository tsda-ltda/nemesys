package models

type RTSMetricInfo struct {
	// PullingTimes is how many times will pull the data.
	PullingTimes int16
	// PullingInterval is the interval between each data request in miliseconds.
	PullingInterval int32
	// Cache duration is the cached data durantion in miliseconds.
	CacheDuration int32
}
