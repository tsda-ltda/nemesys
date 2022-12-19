package alarm

import (
	"context"

	"github.com/fernandotsda/nemesys/shared/logger"
	"github.com/fernandotsda/nemesys/shared/models"
	"github.com/fernandotsda/nemesys/shared/types"
)

func (a *Alarm) notifyAlarm(ctx context.Context, occurency models.AlarmOccurency) {
	profiles, err := a.pg.GetCategoryAlarmProfilesSimplified(ctx, occurency.Category.Id)
	if err != nil {
		a.log.Error("Fail to get category alarm profiles simplified", logger.ErrField(err))
		return
	}
	if len(profiles) == 0 {
		a.log.Debug("Skipping alarm notification, no alarm profile configured")
		return
	}

	var info models.AlarmNotificationInfo

	if occurency.Type == types.ATChecked {
		info, err = a.pg.GetAlarmNotificationInfo(context.Background(),
			occurency.MetricId,
			occurency.ContainerId,
			occurency.Category.Id,
			occurency.ExpressionSimplified.Id,
		)
		info.Expression.Expression = occurency.ExpressionSimplified.Expression
		info.Expression.AlarmCategoryId = occurency.Category.Id
	} else {
		info, err = a.pg.GetAlarmNotificationInfoWitoutExpressions(context.Background(),
			occurency.MetricId,
			occurency.ContainerId,
			occurency.Category.Id,
		)
	}
	if err != nil {
		a.log.Error("Fail to get alarm notification", logger.ErrField(err))
		return
	}

	info.Category.Level = occurency.Category.Level
	info.OccurencyDate = occurency.Time.Unix()

	go a.notifyEmail(info, profiles)
}
