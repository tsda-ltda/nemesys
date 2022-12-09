package alarm

import (
	"context"

	"github.com/fernandotsda/nemesys/shared/logger"
	"github.com/fernandotsda/nemesys/shared/models"
	"github.com/fernandotsda/nemesys/shared/types"
)

func (a *Alarm) notifyAlarm(ctx context.Context, metricAlarmed MetricAlarmed, occurencyDateSeconds int64, alarmType types.AlarmType) {
	profiles, err := a.pg.GetCategoryAlarmProfilesSimplified(ctx, metricAlarmed.Category.Id)
	if err != nil {
		a.log.Error("Fail to get category alarm profiles simplified", logger.ErrField(err))
		return
	}
	if len(profiles) == 0 {
		a.log.Debug("Skipping alarm notification, no alarm profile configured")
		return
	}

	var info models.AlarmNotificationInfo
	if alarmType == types.ATChecked {
		info, err = a.pg.GetAlarmNotificationInfo(context.Background(),
			metricAlarmed.MetricId,
			metricAlarmed.ContainerId,
			metricAlarmed.Category.Id,
			metricAlarmed.ExpressionsSimplified.Id,
		)
		info.Expression.Expression = metricAlarmed.ExpressionsSimplified.Expression
		info.Expression.AlarmCategoryId = metricAlarmed.Category.Id
	} else {
		info, err = a.pg.GetAlarmNotificationInfoWitoutExpressions(context.Background(),
			metricAlarmed.MetricId,
			metricAlarmed.ContainerId,
			metricAlarmed.Category.Id,
		)
	}
	if err != nil {
		a.log.Error("Fail to get alarm notification", logger.ErrField(err))
		return
	}

	info.Category.Level = metricAlarmed.Category.Level
	info.OccurencyDate = occurencyDateSeconds

	go a.notifyEmail(info, profiles, metricAlarmed.Value, alarmType)
}
