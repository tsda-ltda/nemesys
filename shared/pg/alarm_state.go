package pg

import (
	"context"
	"database/sql"

	"github.com/fernandotsda/nemesys/shared/models"
)

const (
	sqlAlarmStateCreate = `INSERT INTO alarm_state (metric_id, state, last_update) VALUES($1, $2, $3);`
	sqlAlarmStateUpdate = `UPDATE alarm_state SET (state, last_update) = ($1, $2) WHERE metric_id = $3;`
	sqlAlarmStateGet    = `SELECT state, last_update FROM alarm_state WHERE metric_id = $1;`
)

func (pg *PG) CreateAlarmState(ctx context.Context, state models.AlarmState) (err error) {
	_, err = pg.db.ExecContext(ctx, sqlAlarmStateCreate, state.MetricId, state.State, state.LastUpdate)
	return err
}

func (pg *PG) UpdateAlarmState(ctx context.Context, state models.AlarmState) (exists bool, err error) {
	t, err := pg.db.ExecContext(ctx, sqlAlarmStateUpdate, state.State, state.LastUpdate, state.MetricId)
	if err != nil {
		return exists, err
	}
	rowsAffected, _ := t.RowsAffected()
	return rowsAffected != 0, nil
}

func (pg *PG) GetAlarmState(ctx context.Context, metricId int64) (exists bool, state models.AlarmState, err error) {
	err = pg.db.QueryRowContext(ctx, sqlAlarmStateGet, metricId).Scan(&state.State, &state.LastUpdate)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, state, nil
		}
		return false, state, err
	}
	state.MetricId = metricId
	return true, state, nil
}
