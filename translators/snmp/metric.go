package snmp

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/fernandotsda/nemesys/shared/models"
	"github.com/fernandotsda/nemesys/shared/types"
)

// Metric is a models.SNMPMetric extention.
type Metric struct {
	models.SNMPMetric

	// Id is the connection id.
	Id int64
	// Type is the metric type.
	Type types.MetricType
	// TTL is the connection time to live.
	TTL time.Duration
	// Ticker is the TTL ticker controller.
	Ticker *time.Ticker
	// OnClose is a callback used when connection is closed.
	OnClose func(c *Metric)
}

func (s *SNMPService) RegisterMetrics(ctx context.Context, req []models.MetricBasicRequestInfo, ttl time.Duration) (metrics []*Metric, err error) {
	ids := make([]int64, len(req))
	for i, m := range req {
		ids[i] = m.Id
	}

	// get metrics configuration
	confs, err := s.pgConn.SNMPv2cMetrics.GetByIds(ctx, ids)
	if err != nil {
		return metrics, err
	}
	metrics = make([]*Metric, len(ids))
	for i, conf := range confs {
		// find type
		var t types.MetricType
		for _, m := range req {
			if m.Id == conf.Id {
				t = m.Type
				break
			}
		}

		metric := &Metric{
			SNMPMetric: conf,
			Type:       t,
			Id:         conf.Id,
			TTL:        ttl,
			Ticker:     time.NewTicker(ttl),
		}

		// run ttl handler
		go metric.RunTTL()
		metric.OnClose = func(m *Metric) {
			// remove connection
			delete(s.metrics, m.Id)
			s.Log.Debug("metric removed, id: " + fmt.Sprint(m.Id))
		}

		// save connection
		s.metrics[conf.Id] = metric
		metrics[i] = metric
	}
	return metrics, err
}

// RegisterMetric register a metric.
func (s *SNMPService) RegisterMetric(ctx context.Context, id int64, t types.MetricType, ttl time.Duration) (metric *Metric, err error) {
	// get metric configuration
	r, err := s.pgConn.SNMPv2cMetrics.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	// check if exists
	if !r.Exists {
		return nil, errors.New("snmp metric not found")
	}

	metric = &Metric{
		SNMPMetric: r.Metric,
		Id:         id,
		Type:       t,
		TTL:        ttl,
		Ticker:     time.NewTicker(ttl),
	}

	// run ttl handler
	go metric.RunTTL()
	metric.OnClose = func(m *Metric) {
		// remove connection
		delete(s.metrics, m.Id)
		s.Log.Debug("metric removed, id: " + fmt.Sprint(m.Id))
	}

	// save connection
	s.metrics[id] = metric
	return metric, nil
}

// Close closes the metric.
func (m *Metric) Close() {
	m.OnClose(m)
}

// Reset TTL ticker. Will panic if called before RunTTL.
func (m *Metric) Reset() {
	m.Ticker.Reset(m.TTL)
}

// RunTTL will set the connection ticker and close in the end.
func (c *Metric) RunTTL() {
	c.Ticker = time.NewTicker(c.TTL)
	defer c.Ticker.Stop()
	for range c.Ticker.C {
		c.Close()
		return
	}
}
