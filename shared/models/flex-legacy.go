package models

import "github.com/fernandotsda/nemesys/shared/types"

type FlexLegacyContainer struct {
	// Id is the unique indentifier.
	Id int32 `json:"id" validate:"-"`
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
	// SerialNumber is the flex serial-number.
	SerialNumber int32 `json:"serial-number" validate:"required"`
	// Model is the flex model.
	Model int16 `json:"model" validate:"required"`
	// City is whitch city the flex is located.
	City string `json:"city" validate:"max=50"`
	// Region is whitch region the flex is located.
	Region string `json:"region" validate:"max=50"`
	// Country is whitch country the flex is located.
	Coutry string `json:"country" validate:"max=50"`
}

type FlexLegacyContainerSNMPConfig struct {
	// CacheDuration is the cache duration in miliseconds of this configuration on the SNMP service.
	CacheDuration int32
	// Target is an ipv4 address.
	Target string
	// Port is a port.
	Port int32
	// Transport is the transport protocol to use ("udp" or "tcp"); if unset "udp" will be used.
	Transport string
	// Community is an SNMP Community string.
	Community string
	// Timeout is the timeout for one SNMP request/response.
	Timeout int32
	// Set the number of retries to attempt.
	Retries int16
	// Max oids per request.
	MaxOids int16
}

type FlexLegacyMetric struct {
	// Id is the metric identifier.
	Id int64 `json:"-" validate:"-"`
	// OID is the snmp object identifier.
	OID string `json:"oid" validate:"required,max=128"`
	// Port is the port flex port.
	Port int16 `json:"port" validate:"-"`
	// PortType is the port type
	PortType int16 `json:"port-type" validate:"-"`
}

type FlexLegacyDatalogDownloadRegistry struct {
	// ContainerId is the container identifier.
	ContainerId int32
	// Metering is the log type Metering.
	Metering int64
	// Status is the log type Status.
	Status int64
	// Command is the log type Command.
	Command int64
	// Virtual is the log type Virtual.
	Virtual int64
}

type FlexLegacyDatalogMetricRequest struct {
	// Id is the metric identifier.
	Id int64
	// Type is the metric type.
	Type types.MetricType
	// DataPolicyId is the data policy id.
	DataPolicyId int16
	// Port is the flex legacy port.
	Port int16
	// PortType is the flex legacy port type.
	PortType types.FlexLegacyPortType
}
