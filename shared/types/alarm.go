package types

type AlarmState int16

const (
	ASNotAlarmed AlarmState = iota
	ASMinorAlarm
	ASMajorAlarm
	ASCriticalAlarm
	ASMAlarmRecognized
)
