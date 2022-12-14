package api

import (
	"context"
	"fmt"

	"github.com/fernandotsda/nemesys/shared/logger"
	"github.com/fernandotsda/nemesys/shared/models"
	"github.com/fernandotsda/nemesys/shared/trap"
)

func (api *API) startTrapListeners() {
	api.trapHandlersMU.Lock()
	defer api.trapHandlersMU.Unlock()

	ctx := context.Background()

	api.Log.Info("Starting trap listeners...")

	listeners, err := api.PG.GetTrapListeners(ctx)
	if err != nil {
		api.Log.Fatal("Fail to get trap listeners on database", logger.ErrField(err))
		return
	}

	api.trapsListeners = make([]*trap.Trap, len(listeners))
	for i, tl := range listeners {
		api.trapsListeners[i] = trap.New(trap.Config{
			Logger:       api.Log,
			TrapListener: tl,
			ServiceIdent: api.GetServiceIdent(),
		})
	}

	api.Log.Info(fmt.Sprintf("Trap listeners started (%d listeners)", len(listeners)))
}

// CreateTrapListener assumes that the trap listener is was created on database
// and run the listener.
func (api *API) CreateTrapListener(tl models.TrapListener) {
	api.trapHandlersMU.Lock()
	defer api.trapHandlersMU.Unlock()

	api.trapsListeners = append(api.trapsListeners, trap.New(trap.Config{
		Logger:       api.Log,
		TrapListener: tl,
		ServiceIdent: api.GetServiceIdent(),
	}))
}

// UpdateTrapListener assumes that the trap listener is was updated on database
// and updates the listener.
func (api *API) UpdateTrapListener(tl models.TrapListener) {
	api.trapHandlersMU.Lock()
	defer api.trapHandlersMU.Unlock()

	for _, _tl := range api.trapsListeners {
		if _tl.GetId() != tl.Id {
			continue
		}
		_tl.Update(tl)
	}
}

// DeleteTrapListener assumes that the trap listener is was deleted from database
// and stops and remove the listener.
func (api *API) DeleteTrapListener(id int32) {
	api.trapHandlersMU.Lock()
	defer api.trapHandlersMU.Unlock()

	listeners := make([]*trap.Trap, len(api.trapsListeners)-1)
	for _, tl := range api.trapsListeners {
		if tl.GetId() == id {
			tl.Close()
			continue
		}
		listeners = append(listeners, tl)
	}
	api.trapsListeners = listeners
}
