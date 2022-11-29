package dhs

import (
	"context"
	"strconv"
	"time"

	"github.com/fernandotsda/nemesys/shared/amqp"
	"github.com/fernandotsda/nemesys/shared/logger"
	"github.com/fernandotsda/nemesys/shared/models"
	"github.com/fernandotsda/nemesys/shared/types"
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

		time := time.Now()

		// write points
		for _, m := range r.Metrics {
			if m.Failed {
				continue
			}

			err = d.influxClient.WritePoint(ctx, models.MetricDataResponse{
				ContainerId:            r.ContainerId,
				MetricBasicDataReponse: m,
			}, time)
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
			for _, p := range d.containersPulling {
				p.Close()
			}
			for _, p := range d.flexsLegacyPulling {
				p.Close()
			}
			for _, w := range d.flexLegacyDatalogWorkers {
				w.Close()
			}
			d.readDatabase()
		case c := <-d.amqph.OnContainerCreated(amqp.QueueDHSContainerCreated):
			if !c.Base.Enabled {
				continue
			}
			d.newFlexLegacyPulling(c.Base.Id)
		case c := <-d.amqph.OnContainerUpdated():
			if !c.Base.Enabled {
				switch c.Base.Type {
				case types.CTFlexLegacy:
					p, ok := d.flexsLegacyPulling[c.Base.Id]
					if !ok {
						continue
					}
					p.Close()
				default:
					for _, p := range d.containersPulling {
						if p.ContainerId == c.Base.Id {
							p.Close()
							continue
						}
					}
				}
			} else {
				switch c.Base.Type {
				case types.CTFlexLegacy:
					_, ok := d.flexsLegacyPulling[c.Base.Id]
					if ok {
						continue
					}
					d.newFlexLegacyPulling(c.Base.Id)
				default:
					continue
				}
			}
		case id := <-d.amqph.OnContainerDeleted():
			for _, dpg := range d.containersPulling {
				if dpg.ContainerId == id {
					dpg.Close()
					continue
				}
			}
			for _, flp := range d.flexsLegacyPulling {
				if flp.id == id {
					flp.Close()
					continue
				}
			}
		case n := <-d.amqph.OnMetricCreated(amqp.QueueDHSMetricCreated):
			if !n.Base.DHSEnabled || !n.Base.Enabled || n.Base.ContainerType == types.CTFlexLegacy {
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
		case <-d.Done():
			return
		}
	}
}
