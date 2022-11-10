package snmp

import (
	"errors"

	"github.com/gosnmp/gosnmp"
)

func ParsePDU(pdu gosnmp.SnmpPDU) (any, error) {
	switch pdu.Type {
	case gosnmp.OctetString:
		b, ok := pdu.Value.([]byte)
		if !ok {
			return nil, errors.New("fail to parse to bytes")
		}
		return string(b), nil
	case gosnmp.Integer:
		return pdu.Value, nil
	default:
		return nil, errors.New("unknown type")
	}
}
