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

	info, err := a.pg.GetAlarmNotificationInfo(context.Background(),
		occurency.MetricId,
		occurency.ContainerId,
		occurency.Category.Id,
	)
	if err != nil {
		a.log.Error("Fail to get alarm notification info", logger.ErrField(err))
		return
	}

	if occurency.Type == types.ATTrapFlexLegacy {
		info.Descr = occurency.TrapDescr
	} else {
		info.Descr = "Alarm occured due to the expression: " + occurency.ExpressionSimplified.Expression
	}

	info.AlarmCategory.Level = occurency.Category.Level
	info.OccurencyDate = occurency.Time.Unix()

	go a.notifyEmail(info, profiles)
	go a.notifyEndpoints(info, profiles)
}
