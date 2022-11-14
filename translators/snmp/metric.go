package snmp

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/fernandotsda/nemesys/shared/models"
	"github.com/fernandotsda/nemesys/shared/types"
)

var ErrMetricNotFound = errors.New("snmp metric not found")

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

func (s *SNMPService) RegisterMetrics(ctx context.Context, req []models.MetricBasicRequestInfo, containerType types.ContainerType, ttl time.Duration) (metrics []*Metric, err error) {
	// get metrics ids
	ids := make([]int64, len(req))
	for i, m := range req {
		ids[i] = m.Id
	}

	// get SNMP config
	confs := make([]models.SNMPMetric, 0, len(req))
	switch containerType {
	case types.CTSNMPv2c:
		confs, err = s.pgConn.SNMPv2cMetrics.GetByIds(ctx, ids)

	case types.CTFlexLegacy:
		confs, err = s.pgConn.FlexLegacyMetrics.GetByIdsAsSNMPMetric(ctx, ids)
	}

	// check errs
	if err != nil {
		return metrics, err
	}

	// create metrics
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
func (s *SNMPService) RegisterMetric(ctx context.Context, request models.MetricRequest, ttl time.Duration) (metric *Metric, err error) {
	var snmpMetric models.SNMPMetric

	// get metric configuration
	switch request.ContainerType {
	case types.CTSNMPv2c:
		r, err := s.pgConn.SNMPv2cMetrics.Get(ctx, request.MetricId)
		if err != nil {
			return nil, err
		}
		if !r.Exists {
			return nil, ErrMetricNotFound
		}

		snmpMetric = r.Metric
	case types.CTFlexLegacy:
		r, err := s.pgConn.FlexLegacyMetrics.GetAsSNMPMetric(ctx, request.MetricId)
		if err != nil {
			return nil, err
		}
		if !r.Exists {
			return nil, ErrMetricNotFound
		}

		snmpMetric = r.Metric
	}

	metric = &Metric{
		SNMPMetric: snmpMetric,
		Id:         request.MetricId,
		Type:       request.MetricType,
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
	s.metrics[request.MetricId] = metric
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
