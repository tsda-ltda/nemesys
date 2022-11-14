package amqp

import "github.com/fernandotsda/nemesys/shared/types"

func GetDataRoutingKey(t types.ContainerType) string {
	switch t {
	case types.CTSNMPv2c:
		return "snmp"
	case types.CTFlexLegacy:
		return "snmp"
	default:
		return "snmp"
	}
}
