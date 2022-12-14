package alarm

import (
	"strconv"

	"github.com/fernandotsda/nemesys/shared/amqp"
	"github.com/fernandotsda/nemesys/shared/amqph"
	"github.com/fernandotsda/nemesys/shared/logger"
	"github.com/fernandotsda/nemesys/shared/models"
	"github.com/fernandotsda/nemesys/shared/types"
)

func (a *Alarm) listenCheckMetricAlarm() {
	var options amqph.ListenerOptions
	options.QueueDeclarationOptions.Name = amqp.QueueAlarmCheckMetricAlarm
	options.QueueBindOptions.Exchange = amqp.ExchangeCheckMetricAlarm

	msgs, done := a.amqph.Listen(options)
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
		case <-done:
			return
		case <-a.Done():
			return
		}
	}
}

func (a *Alarm) listenCheckMetricsAlarm() {
	var options amqph.ListenerOptions
	options.QueueDeclarationOptions.Name = amqp.QueueAlarmCheckMetricsAlarm
	options.QueueBindOptions.Exchange = amqp.ExchangeCheckMetricsAlarm

	msgs, done := a.amqph.Listen(options)
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
		case <-done:
			return
		case <-a.Done():
			return
		}
	}
}

func (a *Alarm) listenMetricAlarmed() {
	direct := strconv.Itoa(int(types.ATDirect))
	trapFlexLegacy := strconv.Itoa(int(types.ATTrapFlexLegacy))

	var options amqph.ListenerOptions
	options.QueueDeclarationOptions.Name = amqp.QueueAlarmMetricAlarmed
	options.QueueBindOptions.Exchange = amqp.ExchangeMetricAlarmed

	msgs, done := a.amqph.Listen(options)
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
		case <-done:
			return
		case <-a.Done():
			return
		}
	}
}

func (a *Alarm) listenMetricsAlarmed() {
	direct := strconv.Itoa(int(types.ATDirect))

	var options amqph.ListenerOptions
	options.QueueDeclarationOptions.Name = amqp.QueueAlarmMetricsAlarmed
	options.QueueBindOptions.Exchange = amqp.ExchangeMetricsAlarmed

	msgs, done := a.amqph.Listen(options)
	for {
		select {
		case d := <-msgs:
			switch d.Type {
			case direct:
				go a.handleDirectMetricsAlarm(d)
			default:
				a.log.Warn("Unsupported amqp message type on metrics alarmed listener, type: " + d.Type)
			}
		case <-done:
			return
		case <-a.Done():
			return
		}
	}
}
