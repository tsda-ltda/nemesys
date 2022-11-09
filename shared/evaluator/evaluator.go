package evaluator

import (
	"context"
	"errors"

	"github.com/Knetic/govaluate"
	"github.com/fernandotsda/nemesys/shared/cache"
	"github.com/fernandotsda/nemesys/shared/db"
	"github.com/fernandotsda/nemesys/shared/types"
)

type Evaluator struct {
	pgConn *db.PgConn
	cache  *cache.Cache
}

func New(pgConn *db.PgConn) *Evaluator {
	return &Evaluator{
		pgConn: pgConn,
		cache:  cache.New(),
	}
}

func (e *Evaluator) Evaluate(v any, metricId int64, mt types.MetricType) (any, error) {
	ctx := context.Background()

	// get on cache
	exists, evexp, err := e.cache.GetMetricEvExpression(ctx, metricId)
	if err != nil {
		return nil, err
	}

	// check if exists
	if !exists {
		// get on database
		exists, evexp, err = e.pgConn.Metrics.GetEvaluableExpression(ctx, metricId)
		if err != nil {
			return nil, err
		}

		// check if exists
		if !exists {
			return nil, errors.New("fail to get evaluable expression, metric does not exists")
		}

		// save cache
		err = e.cache.SetMetricEvExpression(ctx, metricId, evexp)
		if err != nil {
			return nil, err
		}
	}

	// check if is empty
	if evexp == "" {
		return v, nil
	}

	// get expression struct
	exp, err := govaluate.NewEvaluableExpression(evexp)
	if err != nil {
		return nil, err
	}

	// set params
	params := make(map[string]any, 1)
	params["x"] = v

	// evaluate
	r, err := exp.Evaluate(params)
	if err != nil {
		return nil, err
	}

	// parse value
	parsed, err := types.ParseValue(r, mt)
	if err != nil {
		return nil, err
	}

	return parsed, nil

}
