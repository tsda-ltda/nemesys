package dhs

import (
	"context"
	"strconv"
	"time"

	"github.com/fernandotsda/nemesys/shared/amqp"
	"github.com/fernandotsda/nemesys/shared/amqph"
	"github.com/fernandotsda/nemesys/shared/logger"
	"github.com/fernandotsda/nemesys/shared/models"
	"github.com/fernandotsda/nemesys/shared/types"
)

func (d *DHS) metricsDataListener() {
	var options amqph.ListenerOptions
	options.QueueDeclarationOptions.Name = amqp.QueueDHSMetricsDataRes
	options.QueueBindOptions.Exchange = amqp.ExchangeMetricsDataRes
	options.QueueBindOptions.RoutingKey = "dhs"
	msgs, done := d.amqph.Listen(options)

	for {
		select {
		case dv := <-msgs:
			if amqp.ToMessageType(dv.Type) != amqp.OK {
				continue
			}
			ctx := context.Background()

			var r models.MetricsDataResponse
			err := amqp.Decode(dv.Body, &r)
			if err != nil {
				d.log.Error("Fail to decode amqp message body", logger.ErrField(err))
				continue
			}

			time := time.Now()
			for _, m := range r.Metrics {
				if m.Failed {
					continue
				}

				err = d.influxClient.WritePoint(ctx, models.MetricDataResponse{
					ContainerId:            r.ContainerId,
					MetricBasicDataReponse: m,
				}, time)
				if err != nil {
					d.log.Error("Fail to write point", logger.ErrField(err))
					continue
				}
			}
			d.log.Debug("Metrics data saved on influxdb, container id: " + strconv.FormatInt(int64(r.ContainerId), 10))
		case <-done:
			return

		}
	}
}

func (d *DHS) onDataPolicyDeleted(id int16) {
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
}

func (d *DHS) onContainerCreated(base models.BaseContainer, _ any) {
	if !base.Enabled || base.Type != types.CTFlexLegacy {
		return
	}
	d.newFlexLegacyPulling(base.Id)
}

func (d *DHS) onContainerUpdated(base models.BaseContainer, _ any) {
	if !base.Enabled {
		switch base.Type {
		case types.CTFlexLegacy:
			p, ok := d.flexsLegacyPulling[base.Id]
			if !ok {
				return
			}
			p.Close()
		default:
			for _, p := range d.containersPulling {
				if p.ContainerId == base.Id {
					p.Close()
					return
				}
			}
		}
	} else {
		switch base.Type {
		case types.CTFlexLegacy:
			_, ok := d.flexsLegacyPulling[base.Id]
			if ok {
				return
			}
			d.newFlexLegacyPulling(base.Id)
		default:
			return
		}
	}
}

func (d *DHS) onContainerDeleted(id int32) {
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
}

func (d *DHS) onMetricCreated(base models.BaseMetric, _ any) {
	if !base.DHSEnabled || !base.Enabled || !types.IsNonFlex(base.ContainerType) {
		return
	}
	d.AddMetricPulling(models.MetricRequest{
		ContainerId:   base.ContainerId,
		ContainerType: base.ContainerType,
		MetricId:      base.Id,
		MetricType:    base.Type,
		DataPolicyId:  base.DataPolicyId,
	}, time.Second*time.Duration(base.DHSInterval))
}

func (d *DHS) onMetricUpdated(base models.BaseMetric, _ any) {
	if !types.IsNonFlex(base.ContainerType) {
		return
	}

	d.RemoveMetricPulling(base.Id)
	if !base.DHSEnabled || !base.Enabled {
		return
	}

	d.AddMetricPulling(models.MetricRequest{
		ContainerId:   base.ContainerId,
		ContainerType: base.ContainerType,
		MetricId:      base.Id,
		MetricType:    base.Type,
		DataPolicyId:  base.DataPolicyId,
	}, time.Second*time.Duration(base.DHSInterval))
}

func (d *DHS) onMetricDeleted(containerId int32, id int64) {
	d.RemoveMetricPulling(id)
}
