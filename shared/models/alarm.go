package models

import (
	"time"

	"github.com/fernandotsda/nemesys/shared/types"
)

type AlarmProfile struct {
	// Id is the alarm profile unique identifier.
	Id int32 `json:"id" validate:"-"`
	// Name is the alarm profile name.
	Name string `json:"name" validate:"required,max=50"`
	// Descr is the alarm profile description.
	Descr string `json:"descr" validate:"required,max=255"`
}

type AlarmProfileSimplified struct {
	// Id is the alarm profile unique identifier.
	Id int32 `json:"id" validate:"-"`
	// Name is the alarm profile name.
	Name string `json:"name" validate:"required,max=50"`
}

type AlarmProfileEmailWithoutProfileId struct {
	// Id is the alarm profile email id.
	Id int32 `json:"id" validate:"-"`
	// Email is the email.
	Email string `json:"email" validate:"required,max=255"`
}

type AlarmCategory struct {
	// Id is the alarm category unique identifier.
	Id int32 `json:"id" validate:"-"`
	// Name is the alarm category name.
	Name string `json:"name" validate:"required,max=50"`
	// Descr is the alarm category description.
	Descr string `json:"descr" validate:"required,max=255"`
	// Lever is the alarm level.
	Level int32 `json:"level" validate:"-"`
}

type AlarmCategorySimplified struct {
	// Id is the alarm category unique identifier.
	Id int32
	// Lever is the alarm level.
	Level int32
}

type AlarmExpression struct {
	// Id is the alarm expression unique identifier.
	Id int32 `json:"id" validate:"-"`
	// Name is the alarm expression name.
	Name string `json:"name" validate:"required,max=50"`
	// Expression is the alarm expression.
	Expression string `json:"expression" validate:"required,max=255"`
	// AlarmCategoryId is the alarm category id.
	AlarmCategoryId int32 `json:"alarm-category-id" validate:"required"`
}

type AlarmExpressionSimplified struct {
	// Id is the alarm expression unique identifier.
	Id int32 `json:"id" validate:"-"`
	// Expression is the alarm expression.
	Expression string `json:"expression" validate:"required,max=255"`
	// AlarmCategoryId is the alarm category id.
	AlarmCategoryId int32 `json:"alarm-category-id" validate:"required"`
}

type AlarmState struct {
	// MetricId is the metric identifier.
	MetricId int64 `json:"metric-id"`
	// State is the metric state.
	State types.AlarmState `json:"state"`
	// LastUpdate is the last state update in seconds.
	LastUpdate int64 `json:"last-update"`
}

type AlarmNotificationInfo struct {
	// Alarm type is the alarm type.
	AlarmType types.AlarmType
	// MetricId is the metric identifier.
	MetricId int64
	// MetricName is the metric name.
	MetricName string
	// ContainerId is the container id.
	ContainerId int32
	// ContainerName is the container name.
	ContainerName string
	// ContainerType is the container type.
	ContainerType types.ContainerType
	// Category is the alarm category.
	Category AlarmCategory
	// Expression is the alarm expression.
	Expression AlarmExpression
	// OccurrencyDate is the date of occurency in seconds.
	OccurencyDate int64
	// Value is the alarmed value.
	Value any
}

type DirectAlarm struct {
	// MetricId is the metric identifier.
	MetricId int64
	// ContainerId is the container identifier.
	ContainerId int32
	// AlarmCategoryId is the alarm category id.
	AlarmCategoryId int32
	// Value is the alarmed value.
	Value any
}

type FlexLegacyTrapAlarm struct {
	// Timestamp is the trap timestamp.
	Timestamp time.Time
	// Value is the alarmed value.
	Value any
	// PortType is the flex port type
	PortType int16
	// Port is the flex port
	Port int16
	// Description is the alarm description.
	Description string
	// ClientIp is the client ip.
	ClientIp string
	// AlarmCategoryId is the alarm category id.
	AlarmCategoryId int32
}

type TrapCategoryRelation struct {
	// TrapCategoryId is the trap id.
	TrapCategoryId int16 `json:"trap-category-id" validate:"-"`
	// AlarmCategoryId is the category id.
	AlarmCategoryId int32 `json:"alarm-category-id" validate:"-"`
}

type AlarmProfileEmail struct {
	// Id is the alarm profile and email relation id.
	Id int32 `json:"id" validate:"-"`
	// Email is the email.
	Email string `json:"email" validate:"required"`
}

type AlarmOccurency struct {
	// Type is the alarm type.
	Type types.AlarmType
	// MetricId is the metric identifier.
	MetricId int64
	// Time is when the alarm occurency timestamp.
	Time time.Time
	// ContainerId is the container identifier.
	ContainerId int32
	// Category is the alarm category simplified.
	Category AlarmCategorySimplified
	// ExpressionSimplifed is the expression simplified
	// that was used to check the alarm, should only
	// be used if alarm occurency is a producut of
	// a alarm check, not from direct alarms or snmp traps.
	ExpressionSimplified AlarmExpressionSimplified
	// Value is the alarmed value.
	Value any
	// TrapDescr is the trap description, should not be used
	// if alarm occurency was not originated from an snmp trap.
	TrapDescr string
}
