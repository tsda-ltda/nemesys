package rts

import (
	"context"
	"strconv"
	"time"

	"github.com/fernandotsda/nemesys/shared/amqp"
	"github.com/fernandotsda/nemesys/shared/logger"
	"github.com/fernandotsda/nemesys/shared/models"
	"github.com/fernandotsda/nemesys/shared/types"
)

type MetricPulling struct {
	models.RTSMetricConfig
	// Id is the metric identifier.
	Id int64
	// Type is the data type.
	Type types.MetricType
	// pullingTimes is the pulling times.
	pullingTimes int16
	// pullingRemaining is the pulling counter.
	pullingRemaining int16
}

type ContainerPulling struct {
	models.RTSContainerInfo
	// Id is the unique identifier.
	Id int32
	// Type is the container type.
	Type types.ContainerType
	// Metrics is a map of the current metrics pulling.
	Metrics map[int64]*MetricPulling
	// stopCh is the channel to stop the container pulling.
	stopCh chan any
	// rch is the channel for requests.
	rch chan models.AMQPCorrelated[[]byte]
	// RTS is the RTS server.
	RTS *RTS
}

// restartMetricPulling restarts a pulling for a metric if is running, otherwise will do nothing.
func (s *RTS) restartMetricPulling(r models.MetricRequest) {
	// get container pulling
	c, ok := s.pulling[r.ContainerId]
	if !ok {
		return
	}
	m, ok := c.Metrics[r.MetricId]
	if !ok {
		return
	}
	m.Reset()
}

// startMetricPulling starts a pulling for a metric, if the pulling already exists, will restart it.
func (s *RTS) startMetricPulling(r models.MetricRequest, config models.RTSMetricConfig) {
	if config.PullingTimes < 1 {
		return
	}

	go func() {
		s.muStartPulling.Lock()
		defer s.muStartPulling.Unlock()

		// check if container exists
		c, ok := s.pulling[r.ContainerId]
		if !ok {
			e, info, err := s.pgConn.Containers.GetRTSInfo(context.Background(), r.ContainerId)
			if err != nil {
				s.Log.Error("fail to get containers's RTS info", logger.ErrField(err))
				return
			}

			// check if container exists
			if !e {
				s.Log.Warn("fail to start metric pulling, container does not exists")
				return
			}

			// create new container pulling
			c = &ContainerPulling{
				Id:               r.ContainerId,
				Type:             r.ContainerType,
				RTSContainerInfo: info,
				Metrics:          make(map[int64]*MetricPulling),
				stopCh:           make(chan any, 1),
				rch:              s.metricsDataRequestCh,
				RTS:              s,
			}

			// save container
			s.pulling[r.ContainerId] = c

			// run container
			go func(k int32) {
				defer delete(s.pulling, k)
				defer s.Log.Debug("container pulling stoped, id: " + strconv.FormatInt(int64(k), 10))
				c.Run()
			}(r.ContainerId)

			s.Log.Debug("container pulling started, id: " + strconv.FormatInt(int64(r.ContainerId), 10))
		}

		// check if metric already exists in container
		m, ok := c.Metrics[r.MetricId]
		if !ok {
			// push metric to container's metrics
			c.AddMetric(MetricPulling{
				Id:               r.MetricId,
				Type:             r.MetricType,
				RTSMetricConfig:  config,
				pullingRemaining: config.PullingTimes,
				pullingTimes:     config.PullingTimes,
			})
		} else {
			m.Reset()
		}
	}()
}

func (c *ContainerPulling) Run() {
	ticker := time.NewTicker(time.Millisecond * time.Duration(c.PullingInterval))
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			if len(c.Metrics) == 0 {
				c.Stop()
				continue
			}

			r := models.MetricsRequest{
				ContainerId:   c.Id,
				ContainerType: c.Type,
				Metrics:       make([]models.MetricBasicRequestInfo, len(c.Metrics)),
			}

			i := 0
			for k, m := range c.Metrics {
				r.Metrics[i] = models.MetricBasicRequestInfo{
					Id:   m.Id,
					Type: m.Type,
				}

				i++
				m.pullingRemaining--
				if m.pullingRemaining == 0 {
					c.Remove(k)
					continue
				}
				c.Metrics[k] = m
			}

			// encode request
			b, err := amqp.Encode(r)
			if err != nil {
				c.Stop()
				continue
			}

			// send request
			c.rch <- models.AMQPCorrelated[[]byte]{
				CorrelationId: "",
				RoutingKey:    amqp.GetDataRoutingKey(c.Type),
				Info:          b,
			}
		case <-c.stopCh:
			return
		}
	}
}

// AddMetric adds a metric to container's metrics and save the metric info on pending metrics.
func (c *ContainerPulling) AddMetric(m MetricPulling) {
	c.Metrics[m.Id] = &m
}

// Remove removes a metric.
func (c *ContainerPulling) Remove(metricId int64) {
	delete(c.Metrics, metricId)
}

// Stop stops the container pulling.
func (c *ContainerPulling) Stop() {
	c.stopCh <- struct{}{}
}

// Close closes the container pulling.
func (c *ContainerPulling) Close() {
	c.Stop()
}

// Stop sets the remaining pulling times to 0.
func (m *MetricPulling) Stop() {
	m.pullingRemaining = 0
}

// Reset resets the count.
func (m *MetricPulling) Reset() {
	m.pullingRemaining = m.pullingTimes
}
