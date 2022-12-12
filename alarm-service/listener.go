package alarm

import (
	"context"
	"strconv"

	"github.com/fernandotsda/nemesys/shared/amqp"
	"github.com/fernandotsda/nemesys/shared/logger"
	"github.com/fernandotsda/nemesys/shared/models"
	"github.com/fernandotsda/nemesys/shared/types"
)

func (a *Alarm) listenCheckMetricAlarm() {
	msgs, err := a.amqph.Listen(amqp.QueueAlarmCheckMetricAlarm, amqp.ExchangeCheckMetricAlarm)
	if err != nil {
		a.log.Fatal("Fail to listen to metric check alarm", logger.ErrField(err))
		return
	}
	for {
		select {
		case d := <-msgs:
			var dataResponse models.MetricDataResponse
			err := amqp.Decode(d.Body, &dataResponse)
			if err != nil {
				a.log.Error("Fail to decode amqp body", logger.ErrField(err))
				continue
			}
			if dataResponse.Failed {
				continue
			}

			go a.checkMetricAlarm(dataResponse)
		case <-a.Done():
			return
		}
	}
}

func (a *Alarm) listenCheckMetricsAlarm() {
	msgs, err := a.amqph.Listen(amqp.QueueAlarmCheckMetricsAlarm, amqp.ExchangeCheckMetricsAlarm)
	if err != nil {
		a.log.Fatal("Fail to listen to metrics check alarm", logger.ErrField(err))
		return
	}
	for {
		select {
		case d := <-msgs:
			var dataResponse models.MetricsDataResponse
			err := amqp.Decode(d.Body, &dataResponse)
			if err != nil {
				a.log.Error("Fail to decode amqp body", logger.ErrField(err))
				continue
			}

			// remove failed responses to avoid false alarms
			m := make([]models.MetricBasicDataReponse, 0, len(dataResponse.Metrics))
			for _, r := range dataResponse.Metrics {
				if !r.Failed {
					m = append(m, r)
				}
			}
			dataResponse.Metrics = m

			go a.checkMetricsAlarm(dataResponse)
		case <-a.Done():
			return
		}
	}
}

func (a *Alarm) listenMetricAlarmed() {
	msgs, err := a.amqph.Listen(amqp.QueueAlarmMetricAlarmed, amqp.ExchangeMetricAlarmed)
	if err != nil {
		a.log.Fatal("Fail to listen to metric alarmed", logger.ErrField(err))
		return
	}
	for {
		select {
		case d := <-msgs:
			var alarm models.DirectAlarm
			err := amqp.Decode(d.Body, &alarm)
			if err != nil {
				a.log.Error("Fail to decode amqp body", logger.ErrField(err))
				continue
			}

			exists, category, err := a.pg.GetAlarmCategorySimplified(context.Background(), alarm.AlarmCategoryId)
			if err != nil {
				a.log.Error("Fail to get alarm category", logger.ErrField(err))
				continue
			}
			if !exists {
				a.log.Warn("Received metric alarm, but alarm category does not exists, id: " + strconv.Itoa(int(alarm.AlarmCategoryId)))
				continue
			}

			go a.processAlarm(MetricAlarmed{
				MetricId:    alarm.MetricId,
				ContainerId: alarm.ContainerId,
				Category:    category,
				Value:       alarm.Value,
			}, types.ATDirect)
		case <-a.Done():
			return
		}
	}
}

func (a *Alarm) listenMetricsAlarmed() {
	msgs, err := a.amqph.Listen(amqp.QueueAlarmMetricsAlarmed, amqp.ExchangeMetricsAlarmed)
	if err != nil {
		a.log.Fatal("Fail to listen to metrics alarmed", logger.ErrField(err))
		return
	}
	for {
		select {
		case d := <-msgs:
			var alarms []models.DirectAlarm
			err := amqp.Decode(d.Body, &alarms)
			if err != nil {
				a.log.Error("Fail to decode amqp body", logger.ErrField(err))
				continue
			}

			categoriesIds := make([]int32, len(alarms))
			for i, a := range alarms {
				categoriesIds[i] = int32(a.AlarmCategoryId)
			}

			categories, err := a.pg.GetAlarmCategoriesSimplifiedByIds(context.Background(), categoriesIds)
			if err != nil {
				a.log.Error("Fail to get categories ids", logger.ErrField(err))
				continue
			}

			for _, alarm := range alarms {
				for _, c := range categories {
					if c.Id != int32(alarm.AlarmCategoryId) {
						continue
					}

					go a.processAlarm(MetricAlarmed{
						MetricId:    alarm.MetricId,
						ContainerId: alarm.ContainerId,
						Category:    c,
						Value:       alarm.Value,
					}, types.ATDirect)
				}
			}
		case <-a.Done():
			return
		}
	}
}
