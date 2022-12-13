package alarm

import (
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
	direct := strconv.Itoa(int(types.ATDirect))
	trapFlexLegacy := strconv.Itoa(int(types.ATTrapFlexLegacy))

	msgs, err := a.amqph.Listen(amqp.QueueAlarmMetricAlarmed, amqp.ExchangeMetricAlarmed)
	if err != nil {
		a.log.Fatal("Fail to listen to metric alarmed", logger.ErrField(err))
		return
	}
	for {
		select {
		case d := <-msgs:
			switch d.Type {
			case direct:
				go a.handleDirectMetricAlarm(d)
			case trapFlexLegacy:
				go a.handleFlexLegacyTrapAlarm(d)
			default:
				a.log.Warn("Unsupported amqp message type on metrics alarmed listener, type: " + d.Type)
			}
		case <-a.Done():
			return
		}
	}
}

func (a *Alarm) listenMetricsAlarmed() {
	direct := strconv.Itoa(int(types.ATDirect))

	msgs, err := a.amqph.Listen(amqp.QueueAlarmMetricsAlarmed, amqp.ExchangeMetricsAlarmed)
	if err != nil {
		a.log.Fatal("Fail to listen to metrics alarmed", logger.ErrField(err))
		return
	}
	for {
		select {
		case d := <-msgs:
			switch d.Type {
			case direct:
				go a.handleDirectMetricsAlarm(d)
			default:
				a.log.Warn("Unsupported amqp message type on metrics alarmed listener, type: " + d.Type)
			}
		case <-a.Done():
			return
		}
	}
}
