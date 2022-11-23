package dhs

import (
	"context"
	"strconv"
	"time"

	"github.com/fernandotsda/nemesys/shared/amqp"
	"github.com/fernandotsda/nemesys/shared/logger"
	"github.com/fernandotsda/nemesys/shared/models"
)

func (d *DHS) metricsDataListener() {
	msgs, err := d.amqph.Listen(amqp.QueueDHSMetricsDataResponse, amqp.ExchangeMetricsDataResponse,
		models.ListenerOptions{
			Bind: models.QueueBindOptions{
				RoutingKey: "dhs",
			},
		},
	)
	if err != nil {
		d.log.Panic("fail to listen metrics data")
		return
	}
	for dv := range msgs {
		ctx := context.Background()

		if amqp.ToMessageType(dv.Type) != amqp.OK {
			continue
		}

		// decode message
		var r models.MetricsDataResponse
		err = amqp.Decode(dv.Body, &r)
		if err != nil {
			d.log.Error("fail to decode amqp message body", logger.ErrField(err))
			continue
		}

		// write points
		for _, m := range r.Metrics {
			if m.Failed {
				continue
			}

			err = d.influxClient.WritePoint(ctx, models.MetricDataResponse{
				ContainerId:            r.ContainerId,
				MetricBasicDataReponse: m,
			})
			if err != nil {
				d.log.Error("fail to write point", logger.ErrField(err))
				continue
			}
		}
		d.log.Debug("metrics data points saved on influx buffer, container id: " + strconv.FormatInt(int64(r.ContainerId), 10))
	}
}

func (d *DHS) notificationListener() {
	for {
		select {
		case <-d.amqph.OnDataPolicyDeleted():
			for _, dpg := range d.dataPullingGroups {
				dpg.Close()
			}
			d.readDatabase(100, 0)
		case id := <-d.amqph.OnContainerDeleted():
			for k, dpg := range d.dataPullingGroups {
				if dpg.ContainerId == id {
					delete(d.dataPullingGroups, k)
					break
				}
			}
		case n := <-d.amqph.OnMetricCreated():
			if !n.Base.DHSEnabled || !n.Base.Enabled {
				continue
			}
			d.AddMetricPulling(models.MetricRequest{
				ContainerId:   n.Base.ContainerId,
				ContainerType: n.ContainerType,
				MetricId:      n.Base.Id,
				MetricType:    n.Base.Type,
				DataPolicyId:  n.Base.DataPolicyId,
			}, time.Second*time.Duration(n.Base.DHSInterval))
		case n := <-d.amqph.OnMetricUpdated():
			d.RemoveMetricPulling(n.Base.Id)
			if !n.Base.DHSEnabled || !n.Base.Enabled {
				continue
			}

			d.AddMetricPulling(models.MetricRequest{
				ContainerId:   n.Base.ContainerId,
				ContainerType: n.ContainerType,
				MetricId:      n.Base.Id,
				MetricType:    n.Base.Type,
				DataPolicyId:  n.Base.DataPolicyId,
			}, time.Second*time.Duration(n.Base.DHSInterval))
		case p := <-d.amqph.OnMetricDeleted():
			d.RemoveMetricPulling(p.Id)
		}
	}
}
