package models

type TrapListener struct {
	// Id is the trap listener unique identifier.
	Id int32 `json:"id" validate:"-"`
	// Host is the listener host.
	Host string `json:"host" validate:"required"`
	// Port is the listener port.
	Port int32 `json:"port" validate:"required"`
	// AlarmCategoryId is the alarm category id.
	AlarmCategoryId int32 `json:"alarm-category-id" validate:"required"`
	// Community is the snmp trap community.
	Community string `json:"community" validate:"required"`
	// Transport is the snmp transport ("udp" or "tcp").
	Transport string `json:"transport" validate:"min=3"`
}
