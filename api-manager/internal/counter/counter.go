package counter

import (
	"context"
	"sync"
	"time"

	"github.com/fernandotsda/nemesys/shared/influxdb"
	"github.com/fernandotsda/nemesys/shared/logger"
	"github.com/fernandotsda/nemesys/shared/pg"
)

type Counter struct {
	influxClient         *influxdb.Client
	pg                   *pg.PG
	flushTicker          *time.Ticker
	done                 chan any
	requests             int64
	requestsRealtimeData int64
	requestsHistoryData  int64
	Whitelist            sync.Map
	mu                   sync.Mutex
	log                  *logger.Logger
}

func New(influxClient *influxdb.Client, pg *pg.PG, logger *logger.Logger, flushInterval time.Duration) *Counter {
	c := &Counter{
		influxClient: influxClient,
		pg:           pg,
		flushTicker:  time.NewTicker(flushInterval),
		log:          logger,
		Whitelist:    sync.Map{},
		done:         make(chan any),
	}
	c.LoadWhitelist()
	go c.Run()
	return c
}

func (c *Counter) Run() {
	defer c.flushTicker.Stop()
	for {
		select {
		case <-c.flushTicker.C:
			c.mu.Lock()
			r := c.requests
			rrd := c.requestsRealtimeData
			rhd := c.requestsHistoryData
			c.requests = 0
			c.requestsRealtimeData = 0
			c.requestsHistoryData = 0
			c.mu.Unlock()

			if r > 0 {
				c.influxClient.WriteRequestsCount(r)
				c.log.Debug("Requests count writed")
			}
			if rrd > 0 {
				c.influxClient.WriteRealtimeDataRequestsCount(rrd)
				c.log.Debug("Realtime data requests count writed")
			}
			if rhd > 0 {
				c.influxClient.WriteHistoryDataRequestsCount(rhd)
				c.log.Debug("Data history requests count writed")
			}
		case <-c.done:
			return
		}
	}
}

func (c *Counter) IncrRequests() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.requests++
}

func (c *Counter) IncrRealtimeDataRequests() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.requestsRealtimeData++
}

func (c *Counter) IncrDataHistoryRequests() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.requestsHistoryData++
}

func (c *Counter) Close() {
	c.done <- nil
}

func (c *Counter) LoadWhitelist() {
	c.Whitelist.Range(func(key, _ any) bool {
		c.Whitelist.Delete(key)
		return true
	})

	ids, err := c.pg.GetAllCounterWhitelist(context.Background())
	if err != nil {
		c.log.Error("Fail to get all counter whitelist", logger.ErrField(err))
		return
	}

	for _, id := range ids {
		c.Whitelist.Store(id, nil)
	}
}
