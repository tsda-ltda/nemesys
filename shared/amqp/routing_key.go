package amqp

import (
	"errors"

	"github.com/fernandotsda/nemesys/shared/types"
	"github.com/rabbitmq/amqp091-go"
)

var ErrNoRoutingKey = errors.New("no routing key")

func GetDataRoutingKey(t types.ContainerType) (string, error) {
	switch t {
	case types.CTSNMPv2c:
		return "snmp", nil
	case types.CTFlexLegacy:
		return "snmp", nil
	default:
		return "snmp", ErrNoRoutingKey
	}
}

func RouteHeader(serviceName string) amqp091.Table {
	return amqp091.Table{"routing_key": serviceName}
}

func GetRoutingKeyFromHeader(h amqp091.Table) (string, error) {
	rk, ok := h["routing_key"].(string)
	if !ok {
		return "", errors.New("fail to make routing_key assertion from message header")
	}
	return rk, nil
}
