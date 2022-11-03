package models

type SNMPContainer struct {
	// ContainerId is the container id.
	ContainerId int32 `json:"container-id" validate:"-"`

	// CacheDuration is the cache duration in miliseconds of this configuration on the SNMP service.
	CacheDuration int32 `json:"cache-duration" validate:"required"`

	// Target is an ipv4 address.
	Target string `json:"target" validate:"required,max=50"`

	// Port is a port.
	Port int32 `json:"port" validate:"required,max=65535"`

	// Transport is the transport protocol to use ("udp" or "tcp"); if unset "udp" will be used.
	Transport string `json:"transport" validate:"required,max=3"`

	// Community is an SNMP Community string.
	Community string `json:"community" validate:"required,max=50"`

	// Timeout is the timeout for one SNMP request/response.
	Timeout int32 `json:"timeout" validate:"required,max=60000"`

	// Set the number of retries to attempt.
	Retries int16 `json:"retries" validate:"required"`

	// MsgFlags is an SNMPV3 MsgFlags.
	MsgFlags uint8 `json:"msg-flags" validate:"-"`

	// Version is an SNMP Version.
	Version uint8 `json:"version" validate:"min=0,max=3"`

	// Max oids per request.
	MaxOids int16 `json:"max-oids" validate:"required"`
}

type SNMPMetric struct {
	// MetricId is the metric identifier.
	MetricId int64 `json:"-" validate:"-"`
	// OID is the snmp object identifier.
	OID string `json:"oid" validate:"required,max=128"`
}
