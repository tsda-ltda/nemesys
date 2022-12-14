package snmp

import (
	"context"
	"errors"
	"strconv"

	"github.com/fernandotsda/nemesys/shared/amqp"
	"github.com/fernandotsda/nemesys/shared/amqph"
	"github.com/fernandotsda/nemesys/shared/logger"
	"github.com/fernandotsda/nemesys/shared/models"
	"github.com/fernandotsda/nemesys/shared/types"
	"github.com/rabbitmq/amqp091-go"
)

var (
	ErrNoAlarmOID    = errors.New("no correspondent alarm oid")
	ErrNoCategoryOID = errors.New("no correspondent category oid")
)

type flexLegacyAlarm struct {
	MetricId int64
	Value    any
	TrapId   int16
}

func (s *SNMP) notifyAlarms(containerId int32, alarms []flexLegacyAlarm) {
	trapCategoriesIds := make([]int16, len(alarms))
	for i, a := range alarms {
		trapCategoriesIds[i] = a.TrapId
	}

	relations, err := s.pg.GetTrapCategoriesRelationsByIds(context.Background(), trapCategoriesIds)
	if err != nil {
		s.log.Error("Fail to get trap relations by ids on database", logger.ErrField(err))
		return
	}
	if len(relations) == 0 {
		return
	}

	directAlarms := make([]models.DirectAlarm, len(relations))
	for i, rel := range relations {
		for _, a := range alarms {
			if rel.TrapCategoryId == a.TrapId {
				directAlarms[i] = models.DirectAlarm{
					MetricId:        a.MetricId,
					ContainerId:     containerId,
					AlarmCategoryId: rel.AlarmCategoryId,
					Value:           a.Value,
				}
			}
		}
	}

	b, err := amqp.Encode(directAlarms)
	if err != nil {
		s.log.Error("Fail to encode amqp body", logger.ErrField(err))
		return
	}

	s.amqph.Publish(amqph.Publish{
		Exchange: amqp.ExchangeMetricsAlarmed,
		Publishing: amqp091.Publishing{
			Type: strconv.Itoa(int(types.ATDirect)),
			Body: b,
		},
	})
}

func getFlexLegacyAlarmOID(oid string) (alarm string, err error) {
	if len(oid) <= 22 || oid[21:22] != "3" || oid[:1] != "." {
		return "", ErrNoAlarmOID
	}
	portType := oid[23:24]
	switch portType {
	case "2":
		if len(oid) < 31 {
			return alarm, ErrNoAlarmOID
		}
		return ".1.3.6.1.4.1.31957.1.3.2.1.13." + oid[30:], nil
	case "3":
		if len(oid) < 30 {
			return alarm, ErrNoAlarmOID
		}
		return ".1.3.6.1.4.1.31957.1.3.3.1.10." + oid[29:], nil
	case "4":
		if len(oid) < 30 {
			return alarm, ErrNoAlarmOID
		}
		return ".1.3.6.1.4.1.31957.1.3.4.1.11." + oid[29:], nil
	case "6":
		if len(oid) < 31 {
			return alarm, ErrNoAlarmOID
		}
		return ".1.3.6.1.4.1.31957.1.3.6.1.13." + oid[30:], nil
	}
	return "", ErrNoAlarmOID
}

func getFlexLegacyCategoryOID(oid string) (trap string, err error) {
	if len(oid) <= 22 || oid[21:22] != "3" || oid[:1] != "." {
		return "", ErrNoAlarmOID
	}
	portType := oid[23:24]
	switch portType {
	case "2":
		if len(oid) < 31 {
			return trap, ErrNoAlarmOID
		}
		return ".1.3.6.1.4.1.31957.1.3.2.1.12." + oid[30:], nil
	case "3":
		if len(oid) < 30 {
			return trap, ErrNoAlarmOID
		}
		return ".1.3.6.1.4.1.31957.1.3.3.1.8." + oid[29:], nil
	case "4":
		if len(oid) < 30 {
			return trap, ErrNoAlarmOID
		}
		return ".1.3.6.1.4.1.31957.1.3.4.1.8." + oid[29:], nil
	case "6":
		if len(oid) < 31 {
			return trap, ErrNoAlarmOID
		}
		return ".1.3.6.1.4.1.31957.1.3.6.1.12." + oid[30:], nil
	}
	return "", ErrNoAlarmOID
}
