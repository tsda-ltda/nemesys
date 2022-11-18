package evaluator

import (
	"context"
	"errors"

	"github.com/Knetic/govaluate"
	"github.com/fernandotsda/nemesys/shared/cache"
	"github.com/fernandotsda/nemesys/shared/pg"
	"github.com/fernandotsda/nemesys/shared/types"
)

type Evaluator struct {
	pgConn *pg.Conn
	cache  *cache.Cache
}

func New(pgConn *pg.Conn) *Evaluator {
	return &Evaluator{
		pgConn: pgConn,
		cache:  cache.New(),
	}
}

func (e *Evaluator) Evaluate(v any, metricId int64, mt types.MetricType) (any, error) {
	ctx := context.Background()

	var expression string

	// get on cache
	cacheRes, err := e.cache.GetMetricEvExpression(ctx, metricId)
	if err != nil {
		return nil, err
	}

	// check if exists
	if !cacheRes.Exists {
		// get on database
		r, err := e.pgConn.Metrics.GetEvaluableExpression(ctx, metricId)
		if err != nil {
			return nil, err
		}

		// check if exists
		if !r.Exists {
			return nil, errors.New("fail to get evaluable expression, metric does not exists")
		}

		// save cache
		expression = r.Expression
		err = e.cache.SetMetricEvExpression(ctx, metricId, expression)
		if err != nil {
			return nil, err
		}
	}

	// check if is empty
	if expression == "" {
		return v, nil
	}

	// get expression struct
	exp, err := govaluate.NewEvaluableExpression(expression)
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
