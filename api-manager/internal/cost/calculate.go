package cost

import (
	"context"
	"time"

	"github.com/fernandotsda/nemesys/api-manager/internal/api"
	"github.com/fernandotsda/nemesys/shared/models"
)

type additional struct {
	N         int
	BasePlanN int
	Price     float64
}

func calculateAdditional(additionals []additional) (total float64) {
	for _, a := range additionals {
		d := a.N - a.BasePlanN
		if d > 0 {
			total += float64(d) * a.Price
		}
	}
	return total
}

func calculate(ctx context.Context, api *api.API) (result models.ServerCostResult, err error) {
	priceTable, err := api.PG.GetPriceTable(ctx)
	if err != nil {
		return result, err
	}

	basePlan, err := api.PG.GetBasePlan(ctx)
	if err != nil {
		return result, err
	}

	elements, err := api.PG.CountServerElements(ctx)
	if err != nil {
		return result, err
	}

	dps, err := api.PG.GetDataPolicies(ctx)
	if err != nil {
		return result, err
	}

	points, err := api.Influx.CountAllMetricsPoints(dps)
	if err != nil {
		return result, err
	}
	elements.InfluxDataPoints = points

	result.GeneratedAt = time.Now().Unix()
	result.Elements = elements
	result.PriceTable = priceTable
	result.BasePlanCost = basePlan.Cost
	result.TotalCost = basePlan.Cost
	result.AdditionalCost = calculateAdditional([]additional{
		{
			N:         elements.Users,
			BasePlanN: basePlan.Users,
			Price:     priceTable.User,
		},
		{
			N:         elements.Teams,
			BasePlanN: basePlan.Teams,
			Price:     priceTable.Team,
		},
		{
			N:         elements.Contexts,
			BasePlanN: basePlan.Contexts,
			Price:     priceTable.Context,
		},
		{
			N:         elements.ContextualMetrics,
			BasePlanN: basePlan.ContextualMetrics,
			Price:     priceTable.ContextualMetric,
		},
		{
			N:         elements.BasicContainers,
			BasePlanN: basePlan.BasicContainers,
			Price:     priceTable.BasicContainer,
		},
		{
			N:         elements.SNMPv2cContainers,
			BasePlanN: basePlan.SNMPv2cContainers,
			Price:     priceTable.SNMPv2cContainer,
		},
		{
			N:         elements.FlexLegacyContainers,
			BasePlanN: basePlan.FlexLegacyContainers,
			Price:     priceTable.FlexLegacyContainer,
		},
		{
			N:         elements.BasicMetrics,
			BasePlanN: basePlan.BasicMetrics,
			Price:     priceTable.BasicMetric,
		},
		{
			N:         elements.SNMPv2cMetrics,
			BasePlanN: basePlan.SNMPv2cMetrics,
			Price:     priceTable.SNMPv2cMetric,
		},
		{
			N:         elements.FlexLegacyMetrics,
			BasePlanN: basePlan.FlexLegacyMetrics,
			Price:     priceTable.FlexLegacyMetric,
		},
		{
			N:         elements.CustomQueries,
			BasePlanN: basePlan.CustomQueries,
			Price:     priceTable.CustomQuery,
		},
		{
			N:         elements.DataPolicies,
			BasePlanN: basePlan.DataPolicies,
			Price:     priceTable.DataPolicy,
		},
		{
			N:         elements.AlarmExpressions,
			BasePlanN: basePlan.AlarmExpressions,
			Price:     priceTable.AlarmExpression,
		},
		{
			N:         elements.AlarmProfiles,
			BasePlanN: basePlan.AlarmProfiles,
			Price:     priceTable.AlarmProfile,
		},
		{
			N:         elements.Refkeys,
			BasePlanN: basePlan.Refkeys,
			Price:     priceTable.Refkey,
		},
		{
			N:         elements.InfluxDataPoints,
			BasePlanN: basePlan.InfluxDataPoints,
			Price:     priceTable.InfluxDataPoint,
		},
		{
			N:         elements.Requests,
			BasePlanN: basePlan.Requests,
			Price:     priceTable.Request,
		},
		{
			N:         elements.RealtimeDataRequests,
			BasePlanN: basePlan.RealtimeDataRequests,
			Price:     priceTable.HistoryDataRequest,
		},
	})
	result.TotalCost += result.AdditionalCost
	return result, nil
}
