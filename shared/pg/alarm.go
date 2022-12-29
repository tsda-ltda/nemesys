package pg

import (
	"github.com/fernandotsda/nemesys/shared/models"
	"golang.org/x/net/context"
)

const (
	sqlAlarmGetNotficationInfo = `WITH 
		c AS (SELECT name FROM containers WHERE id = $1),
		ca AS (SELECT name FROM alarm_categories WHERE id = $2),
	SELECT name, container_type,
		(SELECT * FROM c),
		(SELECT * FROM ca) FROM metrics WHERE id = $3`
)

func (pg *PG) GetAlarmNotificationInfo(ctx context.Context, metricId int64, containerId int32, categoryId int32) (info models.AlarmNotificationInfo, err error) {
	info.MetricId = metricId
	info.AlarmCategory.Id = categoryId
	info.ContainerId = containerId
	return info, pg.db.QueryRowContext(ctx, sqlAlarmGetNotficationInfo, containerId, categoryId, metricId).Scan(
		&info.MetricName,
		&info.ContainerType,
		&info.ContainerName,
		&info.AlarmCategory.Name,
	)
}
