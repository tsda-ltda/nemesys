package snmp

import (
	"context"
	"strconv"

	"github.com/fernandotsda/nemesys/shared/amqp"
	"github.com/fernandotsda/nemesys/shared/logger"
	"github.com/fernandotsda/nemesys/shared/models"
	"github.com/rabbitmq/amqp091-go"
)

// getMetricListener listen to metric data requests and send to
// the dataProducer.
func (s *SNMPService) getMetricListener() {
	// create socket channel
	ch, err := s.amqp.Channel()
	if err != nil {
		s.Log.Panic("fail to open amqp socket channel", logger.ErrField(err))
		return
	}
	defer ch.Close()

	// declare exchange
	err = ch.ExchangeDeclare(
		amqp.ExchangeGetMetricData, // name
		"direct",                   // type
		true,                       // durable
		false,                      // auto-deleted
		false,                      // internal
		false,                      // no-wait
		nil,                        // arguments
	)
	if err != nil {
		s.Log.Panic("fail to declare exchange", logger.ErrField(err))
		return
	}

	// declare queue
	q, err := ch.QueueDeclare(
		amqp.QueueSNMPGetMetricData, // name
		false,                       // durable
		false,                       // delete when unused
		false,                       // exclusive
		false,                       // no-wait
		nil,                         // arguments
	)
	if err != nil {
		s.Log.Panic("fail to declare queue", logger.ErrField(err))
		return
	}

	// bind queue
	err = ch.QueueBind(
		q.Name,                     // queue name
		"snmp",                     // routing key
		amqp.ExchangeGetMetricData, // exchange
		false,                      // no-wait
		nil,                        // arguments
	)
	if err != nil {
		s.Log.Panic("fail to bind queue", logger.ErrField(err))
		return
	}

	// consume queue
	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	if err != nil {
		s.Log.Panic("fail to consume queue", logger.ErrField(err))
		return
	}

	// close and cancel channels
	closeCh := make(chan *amqp091.Error)
	cancelCh := make(chan string)
	ch.NotifyCancel(cancelCh)
	ch.NotifyClose(closeCh)

	for {
		select {
		case d := <-msgs:
			ctx := context.Background()

			// decode message data
			var r models.MetricRequest
			err = amqp.Decode(d.Body, &r)
			if err != nil {
				s.Log.Error("fail to unmarshal amqp message body", logger.ErrField(err))
				continue
			}
			s.Log.Debug("get metric data request received, metric id: " + strconv.FormatInt(int64(r.MetricId), 10))

			// check if container connection exists
			c, ok := s.conns[r.ContainerId]
			if !ok {
				c, err = s.RegisterAgent(ctx, r.ContainerId, r.ContainerType)
				if err != nil {
					s.Log.Error("fail to register agent", logger.ErrField(err))
					continue
				}
			} else {
				c.Reset()
			}

			// get metric oid
			metric, ok := s.metrics[r.MetricId]
			if !ok {
				metric, err = s.RegisterMetric(ctx, r.MetricId, r.MetricType, c.TTL)
				if err != nil {
					s.Log.Error("fail to register metric", logger.ErrField(err))
					continue
				}
			} else {
				metric.Reset()
				metric.Type = r.MetricType
			}

			// get routing key on message header
			rk, ok := d.Headers["routing_key"].(string)
			if !ok {
				s.Log.Error("fail to make routing_key assertion from message header")
				continue
			}

			// add data request
			s.metricDataReq <- models.AMQPCorrelated[metricRequest]{
				CorrelationId: d.CorrelationId,
				RoutingKey:    rk,
				Info: metricRequest{
					basicMetricRequest: basicMetricRequest{
						SNMPMetric: metric.SNMPMetric,
						Type:       r.MetricType,
					},
					Conn:        c,
					ContainerId: r.ContainerId,
				},
			}
		case err := <-closeCh:
			s.Log.Warn("channel closed", logger.ErrField(err))
			s.Close()
			return
		case r := <-cancelCh:
			s.Log.Warn("channel canceled, reason: " + r)
			s.Close()
			return
		case <-s.stopGetListener:
			return
		}
	}
}

// getMetricsListener listen to metrics data requests and send to
// the metricsDataProducer.
func (s *SNMPService) getMetricsListener() {
	// create socket channel
	ch, err := s.amqp.Channel()
	if err != nil {
		s.Log.Panic("fail to open amqp socket channel", logger.ErrField(err))
		return
	}
	defer ch.Close()

	// declare exchange
	err = ch.ExchangeDeclare(
		amqp.ExchangeGetMetricsData, // name
		"direct",                    // type
		true,                        // durable
		false,                       // auto-deleted
		false,                       // internal
		false,                       // no-wait
		nil,                         // arguments
	)
	if err != nil {
		s.Log.Panic("fail to declare exchange", logger.ErrField(err))
		return
	}

	// declare queue
	q, err := ch.QueueDeclare(
		amqp.QueueSNMPGetMetricsData, // name
		false,                        // durable
		false,                        // delete when unused
		false,                        // exclusive
		false,                        // no-wait
		nil,                          // arguments
	)
	if err != nil {
		s.Log.Panic("fail to declare queue", logger.ErrField(err))
		return
	}

	// bind queue
	err = ch.QueueBind(
		q.Name,                      // queue name
		"snmp",                      // routing key
		amqp.ExchangeGetMetricsData, // exchange
		false,                       // no-wait
		nil,                         // arguments
	)
	if err != nil {
		s.Log.Panic("fail to bind queue", logger.ErrField(err))
		return
	}

	// consume queue
	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	if err != nil {
		s.Log.Panic("fail to consume queue", logger.ErrField(err))
		return
	}

	// close and cancel channels
	closeCh := make(chan *amqp091.Error)
	cancelCh := make(chan string)
	ch.NotifyCancel(cancelCh)
	ch.NotifyClose(closeCh)

	for {
		select {
		case d := <-msgs:
			ctx := context.Background()

			// decode message data
			var r models.MetricsRequest
			err = amqp.Decode(d.Body, &r)
			if err != nil {
				s.Log.Error("fail to unmarshal amqp message body", logger.ErrField(err))
				continue
			}
			s.Log.Debug("get metrics data request received, container id: " + strconv.FormatInt(int64(r.ContainerId), 10))

			// check if agent connection exists
			c, ok := s.conns[r.ContainerId]
			if !ok {
				c, err = s.RegisterAgent(ctx, r.ContainerId, r.ContainerType)
				if err != nil {
					s.Log.Error("fail to register agent", logger.ErrField(err))
					continue
				}
			} else {
				c.Reset()
			}

			// get metrics full information
			metrics := make([]*Metric, 0, len(r.Metrics))
			metricsToRegister := []models.MetricBasicRequestInfo{}

			for _, m := range r.Metrics {
				// check if exists on local map
				metric, ok := s.metrics[m.Id]
				if ok {
					metric.Reset()
					metric.Type = m.Type
					metrics = append(metrics, metric)
					continue
				} else {
					metricsToRegister = append(metricsToRegister, m)
				}
			}

			if len(metricsToRegister) > 0 {
				// register metrics
				newMetrics, err := s.RegisterMetrics(ctx, metricsToRegister, c.TTL)
				if err != nil {
					s.Log.Error("fail to register metrics", logger.ErrField(err))
					continue
				}
				metrics = append(metrics, newMetrics...)
			}

			// get routing key on message header
			rk, ok := d.Headers["routing_key"].(string)
			if !ok {
				s.Log.Error("fail to make routing_key assertion from message header")
				continue
			}

			// create metric requests to fetch data
			reqMetrics := make([]basicMetricRequest, len(metrics))
			for i, m := range metrics {
				reqMetrics[i] = basicMetricRequest{
					SNMPMetric: m.SNMPMetric,
					Type:       m.Type,
				}
			}

			// add data request
			s.metricsDataReq <- models.AMQPCorrelated[metricsRequest]{
				CorrelationId: d.CorrelationId,
				RoutingKey:    rk,
				Info: metricsRequest{
					Conn:        c,
					ContainerId: r.ContainerId,
					Metrics:     reqMetrics,
				},
			}
		case err := <-closeCh:
			s.Log.Warn("channel closed", logger.ErrField(err))
			s.Close()
			return
		case r := <-cancelCh:
			s.Log.Warn("channel canceled, reason: " + r)
			s.Close()
			return
		case <-s.stopGetListener:
			return
		}
	}
}
