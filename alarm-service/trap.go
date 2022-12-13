package alarm

import (
	"context"
	"fmt"
	"strconv"

	"github.com/fernandotsda/nemesys/shared/amqp"
	"github.com/fernandotsda/nemesys/shared/logger"
	"github.com/fernandotsda/nemesys/shared/models"
	"github.com/fernandotsda/nemesys/shared/types"
	"github.com/rabbitmq/amqp091-go"
)

func (a *Alarm) handleFlexLegacyTrapAlarm(d amqp091.Delivery) {
	ctx := context.Background()

	a.log.Debug("Starting flex legacy trap alarm pre-process")
	defer func() {
		a.log.Debug("Flex legacy trap alarm pre-process finished")
	}()

	var trapAlarm models.FlexLegacyTrapAlarm
	err := amqp.Decode(d.Body, &trapAlarm)
	if err != nil {
		a.log.Error("Fail to decode amqp body", logger.ErrField(err))
		return
	}

	exists, containerId, err := a.pg.GetFlexLegacyContainerIdByTargetPort(ctx, trapAlarm.ClientIp)
	if err != nil {
		a.log.Error("Fail to get flex legacy container id by target and port", logger.ErrField(err))
		return
	}
	if !exists {
		a.log.Warn("Fail to handle trap from client ip: " + trapAlarm.ClientIp + ", flex legacy container does not exists")
		return
	}

	exists, metricId, err := a.pg.GetFlexLegacyMetricByPortPortType(ctx, containerId, trapAlarm.Port, trapAlarm.PortType)
	if err != nil {
		a.log.Error("Fail to get metric id by port and port type", logger.ErrField(err))
		return
	}
	if !exists {
		a.log.Warn(fmt.Sprintf("Fail to handle trap from client ip: %s, container id: %d. Metric port: %d and port type: %d, does not exists", trapAlarm.ClientIp, containerId, trapAlarm.Port, trapAlarm.PortType))
		return
	}

	exists, category, err := a.pg.GetAlarmCategorySimplified(ctx, trapAlarm.AlarmCategoryId)
	if err != nil {
		a.log.Error("Fail to get alarm category simplified", logger.ErrField(err))
		return
	}
	if !exists {
		a.log.Warn("Fail to handle flex legacy trap alarm, alarm category does not exists, id: " + strconv.Itoa(int(trapAlarm.AlarmCategoryId)))
		return
	}

	go a.processAlarm(MetricAlarmed{
		MetricId:    metricId,
		ContainerId: containerId,
		Category:    category,
		Value:       trapAlarm.Value,
	}, types.ATTrapFlexLegacy, trapAlarm.Timestamp)
}
