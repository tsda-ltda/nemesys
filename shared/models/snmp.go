package models

import (
	"time"
)

type SNMPAgentConfig struct {
	// TTL is the Time to live of this configuration on the SNMP service memory.
	TTL time.Duration

	// Target is an ipv4 address.
	Target string

	// Port is a port.
	Port uint16

	// Transport is the transport protocol to use ("udp" or "tcp"); if unset "udp" will be used.
	Transport string

	// Community is an SNMP Community string.
	Community string

	// Timeout is the timeout for one SNMP request/response.
	Timeout time.Duration

	// Set the number of retries to attempt.
	Retries int

	// Double timeout in each retry.
	ExponentialTimeout bool

	// MsgFlags is an SNMPV3 MsgFlags.
	MsgFlags uint8

	// Version is an SNMP Version.
	Version uint8

	// Max oids per request.
	MaxOidsPerReq int
}

type SNMPGetMetrics struct {
	// OIDS are the metrics object identifier.
	OIDS []string

	// Target is a ipv4 address.
	Target string

	// Port is the port.
	Port uint16
}

type SNMPGetMetricResult struct {
	OID   string
	Value any
}
