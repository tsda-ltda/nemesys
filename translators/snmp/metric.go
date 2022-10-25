package snmp

import (
	"context"
	"errors"
	"fmt"
	"time"
)

// Metric is the SNMP metric representation.
type Metric struct {
	// Id is the connection id.
	Id int
	// TTL is the connection time to live.
	TTL time.Duration
	// Agent is the GoSNMP configuratation and connection.
	OID string
	// Ticker is the TTL ticker controller.
	Ticker *time.Ticker
	// Closed is the channel to closed the connection.
	Closed chan any
	// OnClose is a callback used when connection is closed.
	OnClose func(c *Metric)
}

// RegisterMetric register a metric.
func (s *SNMPService) RegisterMetric(ctx context.Context, metricId int, ttl time.Duration) error {
	// get metric configuration
	e, conf, err := s.pgConn.SNMPMetrics.Get(ctx, metricId)
	if err != nil {
		return err
	}

	// check if exists
	if !e {
		return errors.New("snmp metric not found")
	}

	m := &Metric{
		Id:     metricId,
		TTL:    ttl,
		OID:    conf.OID,
		Ticker: time.NewTicker(ttl),
	}

	// run ttl handler
	go m.RunTTL()
	m.OnClose = func(m *Metric) {
		// remove connection
		s.metrics[m.Id] = nil
		s.Log.Debug("metric removed, id: " + fmt.Sprint(m.Id))
	}

	// save connection
	s.metrics[metricId] = m
	return nil
}

// Close closes Closed chan.
func (m *Metric) Close() {
	close(m.Closed)
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
	for {
		select {
		case <-c.Closed:
			return
		case <-c.Ticker.C:
			print("a")
			c.Close()
			return
		}
	}
}
