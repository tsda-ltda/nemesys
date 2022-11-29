package snmp

import (
	"context"

	"github.com/fernandotsda/nemesys/shared/models"
	"github.com/fernandotsda/nemesys/shared/types"
)

func (s *SNMP) getSNMPMetrics(request models.MetricsRequest) (metrics []models.SNMPMetric, err error) {
	ctx := context.Background()
	metricIds := make([]int64, len(request.Metrics))
	for i, m := range request.Metrics {
		metricIds[i] = m.Id
	}

	// get metrics on cache
	r, err := s.cache.GetSNMPMetrics(ctx, metricIds)
	if err != nil {
		return metrics, err
	}

	metrics = make([]models.SNMPMetric, 0, len(metricIds))
	notExists := make([]int64, 0)
	for _, res := range r {
		if res.Exists {
			metrics = append(metrics, res.Metric)
			continue
		}
		notExists = append(notExists, res.Metric.Id)
	}

	if len(notExists) > 0 {
		var newMetrics []models.SNMPMetric
		switch request.ContainerType {
		case types.CTSNMPv2c:
			newMetrics, err = s.pg.GetSNMPv2cMetricsByIds(ctx, notExists)
		case types.CTFlexLegacy:
			newMetrics, err = s.pg.FlexLegacyMetricsByIdsAsSNMPMetric(ctx, notExists)
		}
		if err != nil {
			return metrics, err
		}
		metrics = append(metrics, newMetrics...)

		// save on cache
		err = s.cache.SetSNMPMetrics(ctx, newMetrics)
		if err != nil {
			return metrics, err
		}
	}

	return metrics, nil
}
