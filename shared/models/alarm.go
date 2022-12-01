package models

import (
	"github.com/fernandotsda/nemesys/shared/types"
)

type AlarmExpression struct {
	// MetricId is the metric identifier.
	MetricId int64 `json:"metric-id" validate:"-"`
	// MinorExpression is the minor expression.
	MinorExpression string `json:"minor-expression" validate:"max=255"`
	// MajorExpression is the major expression.
	MajorExpression string `json:"major-expression" validate:"max=255"`
	// CriticalExpression is the critical expression.
	CriticalExpression string `json:"critical-expression" validate:"max=255"`
}

type AlarmProfile struct {
	// Id is the alarm profile unique identifier.
	Id int32 `json:"id" validate:"-"`
	// Name is the alarm profile name.
	Name string `json:"name" validate:"required,max=50"`
	// Minor is if the alarm profile wants to receive minor alarms.
	Minor bool `json:"minor" validate:"-"`
	// Major is if the alarm profile wants to receive major alarms.
	Major bool `json:"major" validate:"-"`
	// Critical is if the alarm profile wants to receive critical alarms.
	Critical bool `json:"critical" validate:"-"`
	// Emails are the emails to be notified.
	Emails []string `json:"emails" validate:"required,max=10"`
	// Emails are the wpp numbers to be notified.
	WPP []string `json:"wpp" validate:"required,max=10"`
	// SMS are the sms numbers to be notified.
	SMS []string `json:"sms" validate:"required,max=10"`
	// Telegrams are the telegrams numbers to be notified.
	Telegrams []string `json:"telegrams" validate:"required,max=10"`
}

type AlarmState struct {
	// MetricId is the metric identifier.
	MetricId int64 `json:"metric-id"`
	// State is the current metric state.
	State types.AlarmState `json:"state"`
	// LastMinorTime is the last minor alarm occurency in milliseconds.
	LastMinorTime int64 `json:"last-minor-time"`
	// LastMajorTime is the last major alarm occurency in milliseconds.
	LastMajorTime int64 `json:"last-major-time"`
	// LastMinorTime is the last critical alarm occurency in milliseconds.
	LastCriticalTime int64 `json:"last-critical-time"`
	// LastRecognizationTime is when the metric was recognized.
	LastRecognizationTime int64 `json:"last-recognization-time"`
	// AlwaysAlarmedOnNewAlarm if setted to true will set the state as
	// alarmed every new alarm, ignoring the recognization.
	AlwaysAlarmedOnNewAlarm bool `json:"always-alarmed-on-new-alarm"`
	// RecognizationLifetime is the recognization lifetime in seconds.
	RecognizationLifetime int64 `json:"recognization-lifetime"`
}
