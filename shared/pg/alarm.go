package pg

import (
	"github.com/fernandotsda/nemesys/shared/models"
	"golang.org/x/net/context"
)

const (
	sqlAlarmGetNotficationInfo = `WITH 
		c AS (SELECT name FROM containers WHERE id = $1),
		ca AS (SELECT name FROM alarm_categories WHERE id = $2),
		e AS (SELECT name FROM alarm_expressions WHERE id = $3)
	SELECT name, container_type,
		(SELECT * FROM c),
		(SELECT * FROM ca),
		(select * from e) from metrics where id = $4`
	sqlAlarmGetNotficationInfoWithoutExpression = `WITH 
		c AS (SELECT name FROM containers WHERE id = $1),
		ca AS (SELECT name FROM alarm_categories WHERE id = $2),
	SELECT name, container_type,
		(SELECT * FROM c),
		(SELECT * FROM ca) from metrics where id = $3`
)

func (pg *PG) GetAlarmNotificationInfo(ctx context.Context, metricId int64, containerId int32, categoryId int32, expressionId int32) (info models.AlarmNotificationInfo, err error) {
	info.MetricId = metricId
	info.Category.Id = categoryId
	info.ContainerId = containerId
	info.Expression.Id = expressionId
	return info, pg.db.QueryRowContext(ctx, sqlAlarmGetNotficationInfo, containerId, categoryId, expressionId, metricId).Scan(
		&info.MetricName,
		&info.ContainerType,
		&info.ContainerName,
		&info.Category.Name,
		&info.Expression.Name,
	)
}

func (pg *PG) GetAlarmNotificationInfoWitoutExpressions(ctx context.Context, metricId int64, containerId int32, categoryId int32) (info models.AlarmNotificationInfo, err error) {
	info.MetricId = metricId
	info.Category.Id = categoryId
	info.ContainerId = containerId
	return info, pg.db.QueryRowContext(ctx, sqlAlarmGetNotficationInfoWithoutExpression, containerId, categoryId, metricId).Scan(
		&info.MetricName,
		&info.ContainerType,
		&info.ContainerName,
		&info.Category.Name,
	)
}
