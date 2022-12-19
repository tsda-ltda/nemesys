package alarm

import (
	"context"
	"strconv"

	"github.com/fernandotsda/nemesys/shared/amqp"
	"github.com/fernandotsda/nemesys/shared/logger"
	"github.com/fernandotsda/nemesys/shared/models"
	"github.com/fernandotsda/nemesys/shared/types"
	"github.com/rabbitmq/amqp091-go"
)

func (a *Alarm) handleDirectMetricAlarm(d amqp091.Delivery) {
	var alarm models.DirectAlarm
	err := amqp.Decode(d.Body, &alarm)
	if err != nil {
		a.log.Error("Fail to decode amqp body", logger.ErrField(err))
		return
	}

	exists, category, err := a.pg.GetAlarmCategorySimplified(context.Background(), alarm.AlarmCategoryId)
	if err != nil {
		a.log.Error("Fail to get alarm category", logger.ErrField(err))
		return
	}
	if !exists {
		a.log.Warn("Received metric alarm, but alarm category does not exists, id: " + strconv.Itoa(int(alarm.AlarmCategoryId)))
		return
	}

	a.processAlarm(models.AlarmOccurency{
		MetricId:    alarm.MetricId,
		ContainerId: alarm.ContainerId,
		Category:    category,
		Value:       alarm.Value,
		Type:        types.ATDirect,
		Time:        d.Timestamp,
	})
}

func (a *Alarm) handleDirectMetricsAlarm(d amqp091.Delivery) {
	var alarms []models.DirectAlarm
	err := amqp.Decode(d.Body, &alarms)
	if err != nil {
		a.log.Error("Fail to decode amqp body", logger.ErrField(err))
		return
	}

	categoriesIds := make([]int32, len(alarms))
	for i, a := range alarms {
		categoriesIds[i] = int32(a.AlarmCategoryId)
	}

	categories, err := a.pg.GetAlarmCategoriesSimplifiedByIds(context.Background(), categoriesIds)
	if err != nil {
		a.log.Error("Fail to get categories ids", logger.ErrField(err))
		return
	}

	for _, alarm := range alarms {
		for _, c := range categories {
			if c.Id != int32(alarm.AlarmCategoryId) {
				return
			}

			go a.processAlarm(models.AlarmOccurency{
				MetricId:    alarm.MetricId,
				ContainerId: alarm.ContainerId,
				Category:    c,
				Value:       alarm.Value,
				Type:        types.ATDirect,
				Time:        d.Timestamp,
			})
		}
	}
}
