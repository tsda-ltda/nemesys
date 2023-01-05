package pg

import (
	"context"

	"github.com/fernandotsda/nemesys/shared/models"
)

var AlarmExpressionsValidOrderByColumns = []string{"name", "expression", "category_id"}

type AlarmExpressionQueryFilters struct {
	Name       string `type:"ilike" column:"name"`
	Expression string `type:"ilike" column:"expression"`
	CategoryId int32  `type:"=" column:"level"`
	OrderBy    string
	OrderByFn  string
	Limit      int
	Offset     int
}

func (f AlarmExpressionQueryFilters) GetOrderBy() string {
	return f.OrderBy
}

func (f AlarmExpressionQueryFilters) GetOrderByFn() string {
	return f.OrderByFn
}

func (f AlarmExpressionQueryFilters) GetLimit() int {
	return f.Limit
}

func (f AlarmExpressionQueryFilters) GetOffset() int {
	return f.Offset
}

type AlarmExpressionExistsMetricAndRelationResponse struct {
	// Exists is the alarm expression existence.
	Exists bool
	// MetricExists is the metric existence.
	MetricExists bool
	// RelationExists is the relation existence.
	RelationExists bool
}

const (
	sqlAlarmExpressionsCreate             = `INSERT INTO alarm_expressions (name, expression, category_id) VALUES($1, $2, $3) RETURNING id;`
	sqlAlarmExpressionsUpdate             = `UPDATE alarm_expressions SET (name, expression, category_id) = ($1, $2, $3) WHERE id = $4;`
	sqlAlarmExpressionsDelete             = `DELETE FROM alarm_expressions WHERE id = $1;`
	sqlAlarmExpressionsAddMetric          = `INSERT INTO metrics_alarm_expressions_rel (metric_id, expression_id) VALUES ($1, $2);`
	sqlAlarmExpressionsRemMetric          = `DELETE FROM metrics_alarm_expressions_rel WHERE metric_id = $1 AND expression_id = $2;`
	sqlAlarmExpressionsMetricRelExists    = `SELECT EXISTS (SELECT 1 FROM metrics_alarm_expressions_rel WHERE metric_id = $1 AND expression_id = $2);`
	sqlAlarmExpressionsMetricAndRelExists = `SELECT 
	EXISTS (SELECT 1 FROM alarm_expressions WHERE id = $1),
	EXISTS (SELECT 1 FROM metrics WHERE id = $2),
	EXISTS (SELECT 1 FROM metrics_alarm_expressions_rel WHERE expression_id = $1 AND metric_id = $2);`

	customSqlAlarmExpressionsMGet = `SELECT id, name, expression, category_id FROM alarm_expressions`
)

func (pg *PG) CreateAlarmExpression(ctx context.Context, exp models.AlarmExpression) (id int32, err error) {
	return id, pg.db.QueryRowContext(ctx, sqlAlarmExpressionsCreate, exp.Name, exp.Expression, exp.AlarmCategoryId).Scan(&id)
}

func (pg *PG) UpdateAlarmExpression(ctx context.Context, exp models.AlarmExpression) (exists bool, err error) {
	t, err := pg.db.ExecContext(ctx, sqlAlarmExpressionsUpdate, exp.Name, exp.Expression, exp.AlarmCategoryId, exp.Id)
	if err != nil {
		return exists, err
	}
	rowsAffected, _ := t.RowsAffected()
	return rowsAffected != 0, nil
}

func (pg *PG) DeleteAlarmExpression(ctx context.Context, id int32) (exists bool, err error) {
	t, err := pg.db.ExecContext(ctx, sqlAlarmExpressionsDelete, id)
	if err != nil {
		return exists, err
	}
	rowsAffected, _ := t.RowsAffected()
	return rowsAffected != 0, nil
}

func (pg *PG) GetAlarmExpressions(ctx context.Context, filters AlarmExpressionQueryFilters) (expressions []models.AlarmExpression, err error) {
	sql, params, err := applyFilters(filters, customSqlAlarmExpressionsMGet, AlarmExpressionsValidOrderByColumns)
	if err != nil {
		return nil, err
	}
	rows, err := pg.db.QueryContext(ctx, sql, params...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	expressions = make([]models.AlarmExpression, 0, filters.Limit)
	var exp models.AlarmExpression
	for rows.Next() {
		err = rows.Scan(&exp.Id, &exp.Name, &exp.Expression, &exp.AlarmCategoryId)
		if err != nil {
			return nil, err
		}
		expressions = append(expressions, exp)
	}
	return expressions, nil
}

func (pg *PG) CrateMetricAlarmExpressionRel(ctx context.Context, expressionId int32, metricId int64) (err error) {
	_, err = pg.db.ExecContext(ctx, sqlAlarmExpressionsAddMetric, metricId, expressionId)
	return err
}

func (pg *PG) RemoveMetricAlarmExpressionRel(ctx context.Context, expressionId int32, metricId int64) (exists bool, err error) {
	t, err := pg.db.ExecContext(ctx, sqlAlarmExpressionsRemMetric, metricId, expressionId)
	if err != nil {
		return exists, err
	}
	rowsAffected, _ := t.RowsAffected()
	return rowsAffected != 0, nil
}

func (pg *PG) MetricAlarmExpressionRelExists(ctx context.Context, expressionId int32, metricId int64) (r AlarmExpressionExistsMetricAndRelationResponse, err error) {
	return r, pg.db.QueryRowContext(ctx, sqlAlarmExpressionsMetricAndRelExists, expressionId, metricId).Scan(&r.Exists, &r.MetricExists, &r.RelationExists)
}
