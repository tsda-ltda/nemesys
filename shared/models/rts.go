package models

type RTSMetricInfo struct {
	// PullingTimes is how many times will pull the data.
	PullingTimes int
	// PullingInterval is the interval between each data request in miliseconds.
	PullingInterval int
	// Cache duration is the cached data durantion in miliseconds.
	CacheDuration int
}
