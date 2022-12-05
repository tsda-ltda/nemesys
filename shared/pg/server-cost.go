package pg

import (
	"context"

	"github.com/fernandotsda/nemesys/shared/models"
	"github.com/fernandotsda/nemesys/shared/types"
)

const (
	sqlCostUpdatePriceTable = `UPDATE price_table SET (coin_type, _user, team, context, contextual_metric, basic_container,
		snmpv2c_container, flex_legacy_container, basic_metric, snmpv2c_metric, flex_legacy_metric, custom_query,
		data_policy, alarm_expression, alarm_profile, refkey, api_key, influx_data_point, request, realtime_data_request,
		history_data_request) = ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21) WHERE id = 1;`
	sqlCostGetPriceTable = `SELECT coin_type, _user, team, context, contextual_metric, basic_container,
		snmpv2c_container, flex_legacy_container, basic_metric, snmpv2c_metric, flex_legacy_metric, custom_query,
		data_policy, alarm_expression, alarm_profile, refkey, api_key, influx_data_point, request, realtime_data_request,
		history_data_request FROM price_table WHERE id = 1;`
	sqlCostCountElements = `SELECT 
		(SELECT COUNT (*) FROM users),
		(SELECT COUNT (*) FROM teams),
		(SELECT COUNT (*) FROM contexts),
		(SELECT COUNT (*) FROM contextual_metrics),
		(SELECT COUNT (*) FROM containers WHERE type = $1),
		(SELECT COUNT (*) FROM snmpv2c_containers),
		(SELECT COUNT (*) FROM flex_legacy_containers),
		(SELECT COUNT (*) FROM metrics WHERE container_type = $1),
		(SELECT COUNT (*) FROM snmpv2c_metrics),
		(SELECT COUNT (*) FROM flex_legacy_metrics),
		(SELECT COUNT (*) FROM custom_queries),
		(SELECT COUNT (*) FROM data_policies),
		(SELECT COUNT (*) FROM alarm_expressions),
		(SELECT COUNT (*) FROM alarm_profiles),
		(SELECT COUNT (*) FROM metrics_ref),
		(SELECT COUNT (*) FROM apikeys)`
	sqlCostUpdateBasePlan = `UPDATE base_plan SET (cost, users, teams, contexts, contextual_metrics, basic_containers,
		snmpv2c_containers, flex_legacy_containers, basic_metrics, snmpv2c_metrics, flex_legacy_metrics, custom_queries,
		data_policies, alarm_expressions, alarm_profiles, refkeys, api_keys, influx_data_points, requests, realtime_data_requests,
		history_data_requests) = ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21) WHERE id = 1`
	sqlCostGetBasePlan = `SELECT cost, users, teams, contexts, contextual_metrics, basic_containers,
	snmpv2c_containers, flex_legacy_containers, basic_metrics, snmpv2c_metrics, flex_legacy_metrics, custom_queries,
	data_policies, alarm_expressions, alarm_profiles, refkeys, api_keys, influx_data_points, requests, realtime_data_requests,
	history_data_requests FROM base_plan WHERE id = 1;`
)

func (pg *PG) UpdatePriceTable(ctx context.Context, table models.ServerPriceTable) (err error) {
	_, err = pg.db.ExecContext(ctx, sqlCostUpdatePriceTable,
		table.CoinType,
		table.User,
		table.Team,
		table.Context,
		table.ContextualMetric,
		table.BasicContainer,
		table.SNMPv2cContainer,
		table.FlexLegacyContainer,
		table.BasicMetric,
		table.SNMPv2cMetric,
		table.FlexLegacyMetric,
		table.CustomQuery,
		table.DataPolicy,
		table.AlarmExpression,
		table.AlarmProfile,
		table.Refkey,
		table.APIKey,
		table.InfluxDataPoint,
		table.Request,
		table.RealtimeDataRequest,
		table.HistoryDataRequest,
	)
	return err
}

func (pg *PG) GetPriceTable(ctx context.Context) (table models.ServerPriceTable, err error) {
	return table, pg.db.QueryRowContext(ctx, sqlCostGetPriceTable).Scan(
		&table.CoinType,
		&table.User,
		&table.Team,
		&table.Context,
		&table.ContextualMetric,
		&table.BasicContainer,
		&table.SNMPv2cContainer,
		&table.FlexLegacyContainer,
		&table.BasicMetric,
		&table.SNMPv2cMetric,
		&table.FlexLegacyMetric,
		&table.CustomQuery,
		&table.DataPolicy,
		&table.AlarmExpression,
		&table.AlarmProfile,
		&table.Refkey,
		&table.APIKey,
		&table.InfluxDataPoint,
		&table.Request,
		&table.RealtimeDataRequest,
		&table.HistoryDataRequest,
	)
}

func (pg *PG) CountServerElements(ctx context.Context) (e models.ServerElements, err error) {
	return e, pg.db.QueryRowContext(ctx, sqlCostCountElements, types.CTBasic).Scan(
		&e.Users,
		&e.Teams,
		&e.Contexts,
		&e.ContextualMetrics,
		&e.BasicContainers,
		&e.SNMPv2cContainers,
		&e.FlexLegacyContainers,
		&e.BasicMetrics,
		&e.SNMPv2cMetrics,
		&e.FlexLegacyMetrics,
		&e.CustomQueries,
		&e.DataPolicies,
		&e.AlarmExpressions,
		&e.AlarmProfiles,
		&e.Refkeys,
		&e.APIKeys,
	)
}

func (pg *PG) UpdateBasePlan(ctx context.Context, plan models.ServerBasePlan) (err error) {
	_, err = pg.db.ExecContext(ctx, sqlCostUpdateBasePlan,
		plan.Cost,
		plan.Users,
		plan.Teams,
		plan.Contexts,
		plan.ContextualMetrics,
		plan.BasicContainers,
		plan.SNMPv2cContainers,
		plan.FlexLegacyContainers,
		plan.BasicMetrics,
		plan.SNMPv2cMetrics,
		plan.FlexLegacyMetrics,
		plan.CustomQueries,
		plan.DataPolicies,
		plan.AlarmExpressions,
		plan.AlarmProfiles,
		plan.Refkeys,
		plan.APIKeys,
		plan.InfluxDataPoints,
		plan.Requests,
		plan.RealtimeDataRequests,
		plan.HistoryDataRequests,
	)
	return err
}

func (pg *PG) GetBasePlan(ctx context.Context) (plan models.ServerBasePlan, err error) {
	return plan, pg.db.QueryRowContext(ctx, sqlCostGetBasePlan).Scan(
		&plan.Cost,
		&plan.Users,
		&plan.Teams,
		&plan.Contexts,
		&plan.ContextualMetrics,
		&plan.BasicContainers,
		&plan.SNMPv2cContainers,
		&plan.FlexLegacyContainers,
		&plan.BasicMetrics,
		&plan.SNMPv2cMetrics,
		&plan.FlexLegacyMetrics,
		&plan.CustomQueries,
		&plan.DataPolicies,
		&plan.AlarmExpressions,
		&plan.AlarmProfiles,
		&plan.Refkeys,
		&plan.APIKeys,
		&plan.InfluxDataPoints,
		&plan.Requests,
		&plan.RealtimeDataRequests,
		&plan.HistoryDataRequests,
	)
}
