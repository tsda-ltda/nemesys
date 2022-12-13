package alarm

import (
	"context"
	"strconv"
	"time"

	"github.com/fernandotsda/nemesys/shared/logger"
	"github.com/fernandotsda/nemesys/shared/models"
	"github.com/fernandotsda/nemesys/shared/types"
)

type MetricAlarmed struct {
	MetricId              int64
	ContainerId           int32
	Category              models.AlarmCategorySimplified
	ExpressionsSimplified models.AlarmExpressionSimplified
	Value                 any
}

func (a *Alarm) processAlarm(metricAlarmed MetricAlarmed, alarmType types.AlarmType, timestamp time.Time) {
	ctx := context.Background()

	exists, state, err := a.pg.GetAlarmState(ctx, metricAlarmed.MetricId)
	if err != nil {
		a.log.Error("Fail to get alarm state", logger.ErrField(err))
		return
	}
	state.LastUpdate = timestamp.Unix()

	// notify as soon as possible
	if (exists && state.State == types.ASNotAlarmed) || (!exists) {
		a.notifyAlarm(ctx, metricAlarmed, state.LastUpdate, alarmType)
	}

	if !exists {
		state.MetricId = metricAlarmed.MetricId
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
		a.log.Debug("Alarm state updated, metric id: " + strconv.FormatInt(state.MetricId, 10))
	}

	a.saveAlarmOccurency(metricAlarmed)
	a.log.Debug("Alarm process finished")
}

func (a *Alarm) saveAlarmOccurency(metricAlarmed MetricAlarmed) {
	a.log.Debug("save alarm occurency handler is not implemented yet")
}
