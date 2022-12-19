package alarm

import (
	"context"
	"strconv"

	"github.com/fernandotsda/nemesys/shared/logger"
	"github.com/fernandotsda/nemesys/shared/models"
	"github.com/fernandotsda/nemesys/shared/types"
)

func (a *Alarm) processAlarm(occurency models.AlarmOccurency) {
	ctx := context.Background()
	idString := strconv.FormatInt(occurency.MetricId, 10)

	exists, state, err := a.pg.GetAlarmState(ctx, occurency.MetricId)
	if err != nil {
		a.log.Error("Fail to get alarm state", logger.ErrField(err))
		return
	}

	// notify as soon as possible
	if (exists && state.State == types.ASNotAlarmed) || (!exists) {
		a.notifyAlarm(ctx, occurency)
	}

	state.LastUpdate = occurency.Time.Unix()
	if !exists {
		state.MetricId = occurency.MetricId
		state.State = types.ASAlarmed

		err = a.pg.CreateAlarmState(ctx, state)
		if err != nil {
			a.log.Error("Fail to create alarm state", logger.ErrField(err))
			return
		}
	} else if state.State != types.ASRecognized {
		state.State = types.ASAlarmed
		_, err = a.pg.UpdateAlarmState(ctx, state)
		if err != nil {
			a.log.Error("Fail to update alarm state", logger.ErrField(err))
			return
		}
		a.log.Debug("Alarm state updated, metric id: " + idString)
	}

	a.saveAlarmOccurency(occurency)
	a.log.Debug("Alarm process finished, metric id: " + idString)
}

func (a *Alarm) saveAlarmOccurency(occurency models.AlarmOccurency) {
	a.influxdb.WriteAlarmOccurency(occurency)
	a.log.Debug("Alarm occurency saved on influxdb, metric id: " + strconv.FormatInt(occurency.MetricId, 10))
}
