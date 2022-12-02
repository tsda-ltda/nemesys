package pg

import (
	"context"
	"database/sql"

	"github.com/fernandotsda/nemesys/shared/models"
)

const (
	sqlAlarmExpressionsCreate = `INSERT INTO alarm_expressions (metric_id, minor_expression, major_expression, critical_expression,
		minor_descr, major_descr, critical_descr) VALUES ($1, $2, $3, $4, $5, $6, $7);`
	sqlAlarmExpressionsUpdate = `UPDATE alarm_expressions SET (minor_expression, major_expression, critical_expression, minor_descr, major_descr, critical_descr) 
		= ($1, $2, $3,  $4, $5, $6) WHERE metric_id = $7;`
	sqlAlarmExpressionsDelete = `DELETE FROM alarm_expressions WHERE metric_id = $1;`
	sqlAlarmExpressionsGet    = `SELECT minor_expression, major_expression, critical_expression, minor_descr, major_descr, critical_descr
		FROM alarm_expressions WHERE metric_id = $1;`
	sqlAlarmExpressionsExists = `SELECT 
	EXISTS (SELECT 1 FROM alarm_expressions WHERE metric_id = $1),
	EXISTS (SELECT 1 FROM metrics WHERE id = $1);`
)

func (pg *PG) CreateAlarmExpression(ctx context.Context, alarmExp models.AlarmExpression) (err error) {
	_, err = pg.db.ExecContext(ctx, sqlAlarmExpressionsCreate,
		alarmExp.MetricId,
		alarmExp.MinorExpression,
		alarmExp.MajorExpression,
		alarmExp.CriticalExpression,
		alarmExp.MinorDescr,
		alarmExp.MajorDescr,
		alarmExp.CriticalDescr,
	)
	return err
}

func (pg *PG) UpdateAlarmExpression(ctx context.Context, alarmExp models.AlarmExpression) (exists bool, err error) {
	t, err := pg.db.ExecContext(ctx, sqlAlarmExpressionsUpdate,
		alarmExp.MinorExpression,
		alarmExp.MajorExpression,
		alarmExp.CriticalExpression,
		alarmExp.MinorDescr,
		alarmExp.MajorDescr,
		alarmExp.CriticalDescr,
		alarmExp.MetricId,
	)
	if err != nil {
		return false, err
	}
	rowsAffected, _ := t.RowsAffected()
	return rowsAffected != 0, err
}

func (pg *PG) DeleteAlarmExpression(ctx context.Context, metricId int64) (exists bool, err error) {
	t, err := pg.db.ExecContext(ctx, sqlAlarmExpressionsDelete, metricId)
	if err != nil {
		return false, nil
	}
	rowsAffected, _ := t.RowsAffected()
	return rowsAffected != 0, err
}

func (pg *PG) GetAlarmExpression(ctx context.Context, metricId int64) (exists bool, alarmExp models.AlarmExpression, err error) {
	err = pg.db.QueryRowContext(ctx, sqlAlarmExpressionsGet, metricId).Scan(
		&alarmExp.MinorExpression,
		&alarmExp.MajorExpression,
		&alarmExp.CriticalExpression,
		&alarmExp.MinorDescr,
		&alarmExp.MajorDescr,
		&alarmExp.CriticalDescr,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, alarmExp, nil
		}
		return false, alarmExp, err
	}
	alarmExp.MetricId = metricId
	return true, alarmExp, nil
}

func (pg *PG) AlarmExpressionExists(ctx context.Context, metricId int64) (expressionExists bool, metricExists bool, err error) {
	return expressionExists, metricExists, pg.db.QueryRowContext(ctx, sqlAlarmExpressionsExists, metricId).Scan(&expressionExists, &metricExists)
}
