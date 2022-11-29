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
	pgConn *pg.PG
	cache  *cache.Cache
}

func New(pgConn *pg.PG) *Evaluator {
	return &Evaluator{
		pgConn: pgConn,
		cache:  cache.New(),
	}
}

func (e *Evaluator) Evaluate(v any, metricId int64, mt types.MetricType) (any, error) {
	ctx := context.Background()

	var expression string
	cacheRes, err := e.cache.GetMetricEvExpression(ctx, metricId)
	if err != nil {
		return nil, err
	}

	if !cacheRes.Exists {
		r, err := e.pgConn.GetMetricEvaluableExpression(ctx, metricId)
		if err != nil {
			return nil, err
		}
		if !r.Exists {
			return nil, errors.New("fail to get evaluable expression, metric does not exists")
		}

		expression = r.Expression
		err = e.cache.SetMetricEvExpression(ctx, metricId, expression)
		if err != nil {
			return nil, err
		}
	}

	return evaluate(v, mt, expression)
}

func DirectEvaluation(value any, mt types.MetricType, expression string) (result any, err error) {
	return evaluate(value, mt, expression)
}

func evaluate(value any, mt types.MetricType, expression string) (result any, err error) {
	if expression == "" {
		return value, nil
	}
	exp, err := govaluate.NewEvaluableExpression(expression)
	if err != nil {
		return nil, err
	}
	params := make(map[string]any, 1)
	params["x"] = value
	r, err := exp.Evaluate(params)
	if err != nil {
		return nil, err
	}
	return types.ParseValue(r, mt)
}
