package pg

import (
	"context"

	"github.com/fernandotsda/nemesys/shared/models"
)

const (
	sqlTrapListenersCreate         = `INSERT INTO trap_listeners (host, port, category_id, community, transport) VALUES ($1, $2, $3, $4, $5) RETURNING id;`
	sqlTrapListenersUpdate         = `UPDATE trap_listeners SET (host, port, category_id, community, transport) = ($1, $2, $3, $4, $5) WHERE id = $6;`
	sqlTrapListenersDelete         = `DELETE FROM trap_listeners WHERE id = $1;`
	sqlTrapListenersMGet           = `SELECT id, host, port, category_id, community, transport FROM trap_listeners;`
	sqlTrapListenersHostPortExists = `SELECT EXISTS (SELECT 1 FROM trap_listeners WHERE host = $1 AND port = $2 AND id != $3);`
)

func (pg *PG) CreateTrapListener(ctx context.Context, tl models.TrapListener) (id int32, err error) {
	return id, pg.db.QueryRowContext(ctx, sqlTrapListenersCreate,
		tl.Host,
		tl.Port,
		tl.AlarmCategoryId,
		tl.Community,
		tl.Transport,
	).Scan(&id)
}

func (pg *PG) UpdateTrapListener(ctx context.Context, tl models.TrapListener) (exists bool, err error) {
	t, err := pg.db.ExecContext(ctx, sqlTrapListenersUpdate,
		tl.Host,
		tl.Port,
		tl.AlarmCategoryId,
		tl.Community,
		tl.Transport,
		tl.Id,
	)
	if err != nil {
		return false, err
	}
	rowsAffected, _ := t.RowsAffected()
	return rowsAffected != 0, nil
}

func (pg *PG) DeleteTrapListener(ctx context.Context, id int32) (exists bool, err error) {
	t, err := pg.db.ExecContext(ctx, sqlTrapListenersDelete, id)
	if err != nil {
		return false, err
	}
	rowsAffected, _ := t.RowsAffected()
	return rowsAffected != 0, nil
}

func (pg *PG) GetTrapListeners(ctx context.Context) (listeners []models.TrapListener, err error) {
	rows, err := pg.db.QueryContext(ctx, sqlTrapListenersMGet)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	listeners = []models.TrapListener{}
	for rows.Next() {
		var tl models.TrapListener
		err = rows.Scan(
			&tl.Id,
			&tl.Host,
			&tl.Port,
			&tl.AlarmCategoryId,
			&tl.Community,
			&tl.Transport,
		)
		if err != nil {
			return nil, err
		}
		listeners = append(listeners, tl)
	}
	return listeners, nil
}

func (pg *PG) TrapListenerHostPortExists(ctx context.Context, host string, port int32, id int32) (exists bool, err error) {
	return exists, pg.db.QueryRowContext(ctx, sqlTrapListenersHostPortExists, host, port, id).Scan(&exists)
}
