package rts

import (
	"context"
	"strconv"
	"time"

	"github.com/fernandotsda/nemesys/shared/amqp"
	"github.com/fernandotsda/nemesys/shared/logger"
	"github.com/fernandotsda/nemesys/shared/models"
	"github.com/fernandotsda/nemesys/shared/types"
	"github.com/rabbitmq/amqp091-go"
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
	models.RTSContainerConfig
	// Id is the unique identifier.
	Id int32
	// Type is the container type.
	Type types.ContainerType
	// Metrics is a map of the current metrics pulling.
	Metrics map[int64]*MetricPulling
	// stopCh is the channel to stop the container pulling.
	stopCh chan any
	// RTS is the RTS server.
	RTS     *RTS
	OnClose func(*ContainerPulling)
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
			res, err := s.pgConn.Containers.GetRTSConfig(context.Background(), r.ContainerId)
			if err != nil {
				s.log.Error("fail to get containers's RTS info", logger.ErrField(err))
				return
			}

			// check if container exists
			if !res.Exists {
				s.log.Warn("fail to start metric pulling, container does not exists")
				return
			}

			// create new container pulling
			c = &ContainerPulling{
				Id:                 r.ContainerId,
				Type:               r.ContainerType,
				RTSContainerConfig: res.Config,
				Metrics:            make(map[int64]*MetricPulling),
				stopCh:             make(chan any, 1),
				RTS:                s,
			}

			// save container
			s.pulling[r.ContainerId] = c
			c.OnClose = func(cp *ContainerPulling) {
				delete(s.pulling, cp.Id)
				s.log.Debug("container pulling stoped, id: " + strconv.FormatInt(int64(cp.Id), 10))
			}

			// run container
			go c.Run()

			s.log.Debug("container pulling started, id: " + strconv.FormatInt(int64(r.ContainerId), 10))
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
			if len(c.Metrics) < 1 {
				c.Close()
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
				if m.pullingRemaining < 1 {
					c.remove(k)
					continue
				}
				c.Metrics[k] = m
			}

			// encode request
			b, err := amqp.Encode(r)
			if err != nil {
				c.Close()
				continue
			}

			c.RTS.amqph.PublisherCh <- models.DetailedPublishing{
				Exchange:   amqp.ExchangeMetricsDataRequest,
				RoutingKey: amqp.GetDataRoutingKey(r.ContainerType),
				Publishing: amqp091.Publishing{
					Expiration: amqp.DefaultExp,
					Headers:    amqp.RouteHeader("rts"),
					Body:       b,
				},
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
func (c *ContainerPulling) remove(metricId int64) {
	delete(c.Metrics, metricId)
	c.RTS.log.Debug("metric removed from pulling, metric id: " + strconv.FormatInt(metricId, 10))
}

// Close closes the container pulling.
func (c *ContainerPulling) Close() {
	c.stopCh <- struct{}{}
	c.OnClose(c)
}

// Stop sets the remaining pulling times to 0.
func (m *MetricPulling) Stop() {
	m.pullingRemaining = 0
}

// Reset resets the count.
func (m *MetricPulling) Reset() {
	m.pullingRemaining = m.pullingTimes
}
