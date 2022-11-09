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
	models.RTSMetricInfo
	// Id is the metric identifier.
	Id int64
	// Type is the data type.
	Type types.MetricType
	// pullingTimes is the pulling times.
	pullingTimes int16
	// pullingRemaining is the pulling counter.
	pullingRemaining int16
	OnClose          func(m *MetricPulling)
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

// resetPulling resets a metric pulling.
func (s *RTS) resetPulling(r models.MetricRequest) {
	// get container
	c, ok := s.pulling[r.ContainerId]
	if !ok {
		s.Log.Warn("fail to reset metric pulling, container pulling does not exists")
		return
	}

	// get metric
	m, ok := c.Metrics[r.MetricId]
	if !ok {
		return
	}
	m.UpdateAndReset(r.MetricType)
}

// addPulling starts a pulling for a metric.
func (s *RTS) addPulling(r models.MetricRequest, info models.RTSMetricInfo) {
	if info.PullingTimes < 1 {
		return
	}

	go func() {
		s.muPulling.Lock()
		defer s.muPulling.Unlock()

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
				RTSMetricInfo:    info,
				pullingRemaining: info.PullingTimes,
				pullingTimes:     info.PullingTimes,
			})
		} else {
			m.UpdateAndReset(r.MetricType)
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
	p, ok := c.RTS.pendingMetricsData[c.Id]
	if !ok {
		p = []struct {
			Id int64
			models.RTSMetricInfo
		}{}
	}
	p = append(p, struct {
		Id int64
		models.RTSMetricInfo
	}{
		Id:            m.Id,
		RTSMetricInfo: m.RTSMetricInfo,
	})
	c.RTS.pendingMetricsData[c.Id] = p
}

// Remove removes a metric.
func (c *ContainerPulling) Remove(metricId int64) {
	delete(c.Metrics, metricId)
}

// Stop stops the container pulling.
func (c *ContainerPulling) Stop() {
	for _, m := range c.Metrics {
		m.OnClose(m)
	}
	c.stopCh <- struct{}{}
}

// Close closes the container pulling.
func (c *ContainerPulling) Close() {
	c.Stop()
	delete(c.RTS.pendingMetricsData, c.Id)
}

// UpdateAndReset resets the metric count.
func (m *MetricPulling) UpdateAndReset(t types.MetricType) {
	m.pullingRemaining = m.pullingTimes
	m.Type = t
}
