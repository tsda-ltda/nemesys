package alarm

import (
	"context"
	"strconv"

	"github.com/Knetic/govaluate"
	"github.com/fernandotsda/nemesys/shared/logger"
	"github.com/fernandotsda/nemesys/shared/models"
	"github.com/fernandotsda/nemesys/shared/types"
)

func (a *Alarm) checkMetricAlarm(data models.MetricDataResponse) {
	metricIdString := strconv.FormatInt(data.Id, 10)
	a.log.Debug("Checking metric alarm, metric id: " + metricIdString)
	defer func() {
		a.log.Debug("Metric alarm check process finished, metric id:" + metricIdString)
	}()

	ctx := context.Background()
	var alarmExpressions []models.AlarmExpressionSimplified

	cacheRes, err := a.cache.GetMetricAlarmExpressions(ctx, data.Id)
	if err != nil {
		a.log.Error("Fail to get metric alarm expressions on cache", logger.ErrField(err))
		return
	}
	if cacheRes.Exists {
		alarmExpressions = cacheRes.AlarmExpressions
	} else {
		alarmExpressions, err = a.pg.GetMetricAlarmExpressions(ctx, data.Id)
		if err != nil {
			a.log.Error("Fail to get metric alarm expressions on database", logger.ErrField(err))
			return
		}

		err = a.cache.SetMetricAlarmExpressions(ctx, data.Id, alarmExpressions)
		if err != nil {
			a.log.Error("Fail to set metric alarm expressions on cache", logger.ErrField(err))
			return
		}
	}

	if len(alarmExpressions) == 0 {
		return
	}

	categoriesIds := make([]int32, 0, len(alarmExpressions))
	for _, e := range alarmExpressions {
		var skip bool
		for _, id := range categoriesIds {
			if id == e.AlarmCategoryId {
				skip = true
				break
			}
		}
		if skip {
			continue
		}
		categoriesIds = append(categoriesIds, e.AlarmCategoryId)
	}

	categories, err := a.getCategoriesSimplified(ctx, categoriesIds)
	if err != nil {
		a.log.Error("Fail to get categories simplified", logger.ErrField(err))
		return
	}

	var metricAlarmed MetricAlarmed
	var alarmed bool
	for _, c := range categories {
		for _, e := range alarmExpressions {
			if e.AlarmCategoryId == c.Id {
				alarmed, err = checkAlarm(e.Expression, data.Value)
				if err != nil {
					a.log.Debug("Fail to check metric alarm", logger.ErrField(err))
				}
				if !alarmed {
					continue
				}
				metricAlarmed = MetricAlarmed{
					MetricId:              data.Id,
					ContainerId:           data.ContainerId,
					Category:              c,
					ExpressionsSimplified: e,
					Value:                 data.Value,
				}
				break
			}
		}
		if alarmed {
			break
		}
	}

	if alarmed {
		go a.processAlarm(metricAlarmed, types.ATChecked)
	}
}

func (a *Alarm) checkMetricsAlarm(data models.MetricsDataResponse) {
	containerIdString := strconv.FormatInt(int64(data.ContainerId), 10)
	a.log.Debug("Checking metrics alarm, container id: " + containerIdString)
	defer func() {
		a.log.Debug("Metrics alarm check process finished, container id:" + containerIdString)
	}()
	ctx := context.Background()

	metricIds := make([]int64, len(data.Metrics))
	for i := range metricIds {
		metricIds[i] = data.Metrics[i].Id
	}

	metricsExpressions := make([][]models.AlarmExpressionSimplified, len(data.Metrics))
	cacheRes, err := a.cache.GetMetricsAlarmExpressions(ctx, metricIds)
	if err != nil {
		a.log.Error("Fail to get metrics alarm expressions on cache", logger.ErrField(err))
		return
	}

	remainingIds := make([]int64, 0, len(metricIds))
	for i, r := range cacheRes {
		if r.Exists {
			metricsExpressions[i] = r.AlarmExpressions
			continue
		}
		remainingIds = append(remainingIds, metricIds[i])
	}

	if len(remainingIds) > 0 {
		remainingExpressions, err := a.pg.GetMetricsAlarmExpressions(ctx, remainingIds)
		if err != nil {
			a.log.Error("Fail to get metrics alarm expressions", logger.ErrField(err))
			return
		}

		err = a.cache.SetMetricsAlarmExpressions(ctx, remainingIds, remainingExpressions)
		if err != nil {
			a.log.Error("Fail to set metrics alarm expressions on cache", logger.ErrField(err))
			return
		}

		for i, id := range metricIds {
			for j, rId := range remainingIds {
				if id == rId {
					metricsExpressions[i] = remainingExpressions[j]
				}
			}
		}
	}

	categoriesIds := make([]int32, 0, len(metricsExpressions))
	for _, me := range metricsExpressions {
		for _, e := range me {
			var founded bool
			for _, id := range categoriesIds {
				if id == e.AlarmCategoryId {
					founded = true
					break
				}
			}
			if !founded {
				categoriesIds = append(categoriesIds, e.AlarmCategoryId)
			}
		}
	}

	categories, err := a.getCategoriesSimplified(ctx, categoriesIds)
	if err != nil {
		a.log.Error("Fail to get categories simplified", logger.ErrField(err))
		return
	}

	for _, c := range categories {
		for i, me := range metricsExpressions {
			for _, e := range me {
				if c.Id != e.AlarmCategoryId {
					continue
				}

				alarmed, err := checkAlarm(e.Expression, data.Metrics[i].Value)
				if err != nil {
					a.log.Debug("Fail to check metric alarm", logger.ErrField(err))
					continue
				}

				if !alarmed {
					continue
				}

				go a.processAlarm(MetricAlarmed{
					MetricId:              metricIds[i],
					ContainerId:           data.ContainerId,
					Category:              c,
					ExpressionsSimplified: e,
					Value:                 data.Metrics[i].Value,
				}, types.ATChecked)
			}
		}
	}
}

func checkAlarm(expression string, value any) (alarmed bool, err error) {
	if expression == "" {
		return alarmed, nil
	}
	exp, err := govaluate.NewEvaluableExpression(expression)
	if err != nil {
		return alarmed, ErrInvalidExpression
	}
	params := make(map[string]any, 1)
	params["x"] = value
	r, err := exp.Evaluate(params)
	if err != nil {
		return alarmed, ErrInvalidExpression
	}
	alarmed, ok := r.(bool)
	if !ok {
		return alarmed, ErrInvalidExpression
	}
	return alarmed, nil
}
