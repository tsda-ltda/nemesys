package models

import (
	"time"

	"github.com/gosnmp/gosnmp"
)

type SNMPMetric struct {
	// Id is the metric identifier.
	Id int64 `json:"-" validate:"-"`
	// OID is the snmp object identifier.
	OID string `json:"oid" validate:"required,max=128"`
}

type SNMPAgent struct {
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
	// MsgFlags is an SNMPV3 MsgFlags.
	MsgFlags gosnmp.SnmpV3MsgFlags
	// SecurityModel is an SNMPV3 Security Model.
	SecurityModel gosnmp.SnmpV3SecurityModel
	// SecurityParameters is an SNMPV3 Security Model parameters struct.
	SecurityParameters gosnmp.SnmpV3SecurityParameters
	// ContextEngineID is SNMPV3 ContextEngineID in ScopedPDU.
	ContextEngineID string
	// ContextName is SNMPV3 ContextName in ScopedPDU
	ContextName string
}
