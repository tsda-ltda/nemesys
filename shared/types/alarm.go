package types

type AlarmState int16
type AlarmType uint8

const (
	ASNotAlarmed AlarmState = iota
	ASAlarmed
	ASRecognized
)

const (
	// ATDirect is all direct alarm notifications generated by any service,
	// like RTS fetching an alarm value of a flex port and reporting the
	// alarm directly.
	ATDirect AlarmType = iota
	// ATChecked is all alarms generated by the metric data alarm check
	// process.
	ATChecked
	// ATTrap is all alarms received by traps.
	ATTrap
)
