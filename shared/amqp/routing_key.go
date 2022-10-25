package amqp

import "github.com/fernandotsda/nemesys/shared/types"

func GetDataRoutingKey(t types.ContainerType) string {
	switch t {
	case types.CTSNMP:
		return "snmp"
	default:
		return "snmp"
	}
}
