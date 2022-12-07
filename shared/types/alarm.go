package types

type AlarmState int16
type AlarmType uint8

const (
	ASNotAlarmed AlarmState = iota
	ASAlarmed
	ASMinorAlarm
	ASMajorAlarm
	ASCriticalAlarm
	ASMAlarmRecognized
)

const (
	ATNotAlarmed AlarmType = iota
	ATAlarmed
	ATMinorAlarm
	ATMajorAlarm
	ATCriticalAlarm
)

func IsAlarmed(at AlarmType) bool {
	return at > ATNotAlarmed
}
