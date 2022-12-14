package dhs

import (
	"strconv"
	"time"

	"github.com/fernandotsda/nemesys/shared/amqp"
	"github.com/fernandotsda/nemesys/shared/amqph"
	"github.com/fernandotsda/nemesys/shared/logger"
	"github.com/fernandotsda/nemesys/shared/models"
	"github.com/fernandotsda/nemesys/shared/types"
	"github.com/rabbitmq/amqp091-go"
)

// getPullingGroupKey returns a pulling group map key.
func getPullingGroupKey(containerId int32, interval time.Duration) string {
	return strconv.FormatInt(int64(containerId), 10) + "_" + strconv.FormatInt(interval.Milliseconds(), 10)
}

// containerPulling is desgined for non-flex containers to fetch data periodically.
type containerPulling struct {
	// ContainerId is the container id.
	ContainerId int32
	// containerIdString is the container id formated as string.
	containerIdString string
	// Container type is the container type.
	ContainerType types.ContainerType
	// encodedRequest is the metrics request encoded.
	encodedRequest []byte
	// ticker is the pulling ticker.
	ticker *time.Ticker
	// Interval is the pulling interval.
	interval time.Duration
	// close is the close channel.
	close chan any
	// OnClose is the callback called on close.
	OnClose func(*containerPulling)
	// DHS is the Data History service pointer.
	dhs *DHS
	// nMetrics is the number of metrics running on this group.
	nMetrics int
}

func (c *containerPulling) Run() {
	defer c.ticker.Stop()
	for {
		select {
		case <-c.ticker.C:
			if !c.dhs.IsReady {
				continue
			}
			if c.nMetrics == 0 {
				c.dhs.log.Debug("Skipping pulling, encoded request is empty, container id: " + c.containerIdString)
				continue
			}

			routingKey, err := amqp.GetDataRoutingKey(c.ContainerType)
			if err != nil {
				continue
			}

			c.dhs.amqph.Publish(amqph.Publish{
				Exchange:   amqp.ExchangeMetricsDataReq,
				RoutingKey: routingKey,
				Publishing: amqp091.Publishing{
					Headers:    amqp.RouteHeader("dhs"),
					Expiration: amqp.DefaultExp,
					Body:       c.encodedRequest,
				},
			})

			c.dhs.log.Debug("Metrics data request sent, container id: " + c.containerIdString)
		case <-c.close:
			return
		}
	}
}

// AddMetricPulling adds a metric to a DataPullingGroup. If DataPullingGroup does not exists, will create a new one.
func (d *DHS) AddMetricPulling(info models.MetricRequest, interval time.Duration) {
	key := getPullingGroupKey(info.ContainerId, interval)
	group, ok := d.containersPulling[key]
	if !ok {
		var err error
		group, err = d.CreatePullingGroup(info.ContainerId, info.ContainerType, interval)
		if err != nil {
			d.log.Error("Fail to create pulling metric", logger.ErrField(err))
			return
		}
	}
	group.AddMetric(models.MetricBasicRequestInfo{
		Id:           info.MetricId,
		Type:         info.MetricType,
		DataPolicyId: info.DataPolicyId,
	})
}

// RemoveMetricPulling removes a metric pulling if it exists.
func (d *DHS) RemoveMetricPulling(metricId int64) {
	key, ok := d.metricsContainerMap[metricId]
	if !ok {
		return
	}
	group, ok := d.containersPulling[key]
	if !ok {
		d.log.Error("Fail to remove metric pulling, metric id have a correspondent group key, but group does not exists.")
		return
	}
	group.RemoveMetric(metricId)
}

// CreatePullingGroup creates a new DataPullingGroup.
func (d *DHS) CreatePullingGroup(containerId int32, containerType types.ContainerType, interval time.Duration) (*containerPulling, error) {
	// create and encode request
	request := models.MetricsRequest{
		ContainerId:   containerId,
		ContainerType: containerType,
		Metrics:       []models.MetricBasicRequestInfo{},
	}
	b, err := amqp.Encode(request)
	if err != nil {
		return nil, err
	}

	group := containerPulling{
		ticker:            time.NewTicker(interval),
		interval:          interval,
		ContainerId:       containerId,
		ContainerType:     containerType,
		close:             make(chan any),
		dhs:               d,
		encodedRequest:    b,
		containerIdString: strconv.FormatInt(int64(containerId), 10),
		OnClose: func(g *containerPulling) {
			delete(d.containersPulling, getPullingGroupKey(containerId, interval))
		},
	}

	key := getPullingGroupKey(containerId, interval)
	d.containersPulling[key] = &group
	go group.Run()
	return &group, nil
}

// AddMetric adds a metric.
func (g *containerPulling) AddMetric(info models.MetricBasicRequestInfo) {
	// decode current request
	var request models.MetricsRequest
	err := amqp.Decode(g.encodedRequest, &request)
	if err != nil {
		g.dhs.log.Error("Fail to decode encoded request", logger.ErrField(err))
		return
	}

	request.Metrics = append(request.Metrics, info)

	// encode request
	b, err := amqp.Encode(request)
	if err != nil {
		g.dhs.log.Error("Fail to encode metrics request", logger.ErrField(err))
		return
	}
	g.encodedRequest = b
	g.dhs.metricsContainerMap[info.Id] = getPullingGroupKey(g.ContainerId, g.interval)
	g.nMetrics++
}

func (c *containerPulling) RemoveMetric(metricId int64) {
	// decode current request
	var request models.MetricsRequest
	err := amqp.Decode(c.encodedRequest, &request)
	if err != nil {
		c.dhs.log.Error("Fail to decode encoded request", logger.ErrField(err))
		return
	}

	// remove metric
	metrics := make([]models.MetricBasicRequestInfo, len(request.Metrics)-1)
	for _, m := range request.Metrics {
		if m.Id != metricId {
			metrics = append(metrics, m)
		}
	}

	// check if no metrics remain
	if len(metrics) == 0 {
		c.Close()
		return
	}

	// save metrics
	request.Metrics = metrics

	// encode request
	b, err := amqp.Encode(request)
	if err != nil {
		c.dhs.log.Error("Fail to encode metrics request", logger.ErrField(err))
		return
	}
	c.encodedRequest = b
	delete(c.dhs.metricsContainerMap, metricId)
	c.nMetrics--
}

// Interval returns the current ticker interval.
func (c *containerPulling) Interval() time.Duration {
	return c.interval
}

func (c *containerPulling) Close() {
	c.close <- nil
	c.OnClose(c)
}
