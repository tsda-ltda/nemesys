package dhs

import (
	"strconv"
	"time"

	"github.com/fernandotsda/nemesys/shared/env"
	"github.com/fernandotsda/nemesys/shared/logger"
)

type flexLegacyPulling struct {
	// id is the flex legacy container identifier.
	id int32
	// ticket is the pulling ticker.
	ticker *time.Ticker
	// closed is closed chan.
	closed chan any
	// onPull is the handler called every pull.
	onPull func(containerId int32)
	// onClose is the callback called on close.
	onClose func(containerId int32)
}

func (d *DHS) newFlexLegacyPulling(containerId int32) {
	interval, err := strconv.ParseInt(env.DHSFlexLegacyDatlogRequestInterval, 0, 64)
	if err != nil {
		d.log.Fatal("Fail to parse env.DHSFlexLegacyDatlogRequestInterval", logger.ErrField(err))
		return
	}

	f := &flexLegacyPulling{
		ticker: time.NewTicker(time.Minute * (time.Duration(interval) / 2)),
		id:     containerId,
		closed: make(chan any),
		onPull: d.getFlexLegacyDatalog,
	}
	go f.Run()
	d.flexsLegacyPulling[containerId] = f
	f.onClose = func(containerId int32) {
		delete(d.flexsLegacyPulling, containerId)
		d.log.Debug("FlexLegacy pulling removed, container id: " + strconv.FormatInt(int64(containerId), 10))
	}
	d.log.Debug("FlexLegacy pulling added, container id: " + strconv.FormatInt(int64(containerId), 10))
}

func (f *flexLegacyPulling) Run() {
	defer f.ticker.Stop()
	for {
		select {
		case <-f.ticker.C:
			f.onPull(f.id)
		case <-f.closed:
			return
		}
	}
}

func (f *flexLegacyPulling) Close() {
	f.closed <- nil
	f.onClose(f.id)
}
