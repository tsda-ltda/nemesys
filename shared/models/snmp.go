package models

type SNMPMetric struct {
	// Id is the metric identifier.
	Id int64 `json:"-" validate:"-"`
	// OID is the snmp object identifier.
	OID string `json:"oid" validate:"required,max=128"`
}
