package models

import (
	"time"

	"github.com/gosnmp/gosnmp"
)

type SNMPv2cContainer struct {
	// Id is the container id.
	Id int32 `json:"container-id" validate:"-"`

	// Target is an ipv4 address.
	Target string `json:"target" validate:"required,max=15"`

	// Port is a port.
	Port int32 `json:"port" validate:"required,max=65535"`

	// Transport is the transport protocol to use ("udp" or "tcp"); if unset "udp" will be used.
	Transport string `json:"transport" validate:"required,max=3"`

	// Community is an SNMP Community string.
	Community string `json:"community" validate:"required,max=50"`

	// Timeout is the timeout for one SNMP request/response.
	Timeout int32 `json:"timeout" validate:"required,min=100,max=60000"`

	// Set the number of retries to attempt.
	Retries int16 `json:"retries" validate:"required"`

	// Max oids per request.
	MaxOids int16 `json:"max-oids" validate:"required"`
}

type SNMPv2cAgent struct {
	// Target is an ipv4 address.
	Target string
	// Port is a port.
	Port uint16
	// Transport is the transport protocol to use ("udp" or "tcp"); if unset "udp" will be used.
	Transport string
	// Community is an SNMP Community string.
	Community string
	// Version is an SNMP Version.
	Version gosnmp.SnmpVersion
	// Timeout is the timeout for one SNMP request/response.
	Timeout time.Duration
	// Set the number of retries to attempt.
	Retries int
	// MaxOids is the maximum number of oids allowed in a Get().
	// (default: MaxOids)
	MaxOids int
}
