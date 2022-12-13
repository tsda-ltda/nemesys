package trap

import (
	"errors"
	"net"
	"strconv"
	"time"

	"github.com/fernandotsda/nemesys/shared/amqp"
	"github.com/fernandotsda/nemesys/shared/logger"
	"github.com/fernandotsda/nemesys/shared/models"
	"github.com/fernandotsda/nemesys/shared/types"
	g "github.com/gosnmp/gosnmp"
	"github.com/rabbitmq/amqp091-go"
)

const (
	oidTimestamp = ".1.3.6.1.2.1.1.3.0"
	oidOidValue  = ".1.3.6.1.6.3.1.1.4.1.0"
)

var (
	ErrInvalidTimestampValue   = errors.New("invalid trap timestamp value")
	ErrInvalidPortValueOID     = errors.New("invalid port value oid")
	ErrInvalidPortValue        = errors.New("invalid trap port value")
	ErrInvalidDescriptionValue = errors.New("invalid trap description value")
	ErrInvalidTimestampOID     = errors.New("invalid timestamp variable")
	ErrInvalidOIDValueOID      = errors.New("invalid oid value oid")
	ErrInvalidVariablesLength  = errors.New("invalid trap binding variables length")
)

func (t *Trap) handler(s *g.SnmpPacket, u *net.UDPAddr) {
	if s.Community != t.tl.Community || s.Error != g.NoError {
		t.log.Debug("Trap package dropped")
		return
	}

	t.handleFlexLegacyTrap(s, u)
}

func (t *Trap) handleFlexLegacyTrap(s *g.SnmpPacket, u *net.UDPAddr) {
	tp, err := parseFlexLegacyTrapVariables(s.Variables)
	if err != nil {
		t.log.Warn("Fail to parse flex legacy trap variables", logger.ErrField(err))
		return
	}

	tp.ClientIp = u.IP.String()
	tp.AlarmCategoryId = t.tl.AlarmCategoryId

	b, err := amqp.Encode(tp)
	if err != nil {
		t.log.Error("Fail to encode amqpm body", logger.ErrField(err))
		return
	}

	t.amqph.PublisherCh <- models.DetailedPublishing{
		Exchange: amqp.ExchangeMetricAlarmed,
		Publishing: amqp091.Publishing{
			Type: strconv.Itoa(int(types.ATTrapFlexLegacy)),
			Body: b,
		},
	}
}

func parseFlexLegacyTrapVariables(variables []g.SnmpPDU) (tp models.FlexLegacyTrapAlarm, err error) {
	if len(variables) < 5 {
		return tp, ErrInvalidVariablesLength
	}
	if len(variables[3].Name) < 25 {
		return tp, ErrInvalidPortValueOID
	}
	if variables[0].Name != oidTimestamp {
		return tp, ErrInvalidTimestampOID
	}
	if variables[1].Name != oidOidValue {
		return tp, ErrInvalidOIDValueOID
	}

	timestamp, err := types.ParseValue(variables[0].Value, types.MTInt)
	if err != nil {
		return tp, ErrInvalidTimestampValue
	}
	descrBytes, ok := variables[3].Value.([]byte)
	if !ok {
		return tp, ErrInvalidDescriptionValue
	}
	port, err := types.ParseValue(variables[4].Value, types.MTInt)
	if err != nil {
		return tp, ErrInvalidPortValue
	}
	portType, err := types.ParseValue(variables[2].Name[23:24], types.MTInt)
	if err != nil {
		return tp, ErrInvalidPortValueOID
	}

	tp = models.FlexLegacyTrapAlarm{
		Timestamp:   time.Unix(timestamp.(int64), 0),
		Value:       variables[2].Value,
		PortType:    int16(portType.(int64)),
		Port:        int16(port.(int64)),
		Description: string(descrBytes),
	}

	return tp, nil
}
