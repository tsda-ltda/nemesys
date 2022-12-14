package rts

import (
	"context"
	"errors"

	"github.com/fernandotsda/nemesys/shared/models"
)

func (s *RTS) getRTSMetricConfig(ctx context.Context, metricId int64) (cfg models.RTSMetricConfig, err error) {
	cacheRes, err := s.cache.GetRTSMetricConfig(ctx, metricId)
	if err != nil {
		return cfg, err
	}
	if cacheRes.Exists {
		return cacheRes.Config, nil
	}

	exists, conf, err := s.pg.GetMetricRTSConfig(ctx, metricId)
	if err != nil {
		return cfg, err
	}
	if !exists {
		return cfg, errors.New("metric realtime service configuration does not exists")
	}
	return conf, nil
}
