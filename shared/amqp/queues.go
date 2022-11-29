package amqp

const (
	QueueSNMPMetricDataRequest  = "services.snmp.metric.data_request"
	QueueSNMPMetricsDataRequest = "services.snmp.metrics.data_request"
	QueueRTSMetricDataRequest   = "services.rts.metric.data_request"
	QueueRTSMetricDataResponse  = "services.rts.metric.data_response"
	QueueRTSMetricsDataResponse = "services.rts.metrics.data_response"
	QueueDHSMetricsDataResponse = "services.dhs.metrics.data_response"
	QueueDHSMetricCreated       = "services.dhs.notification.metric.create"
	QueueDHSContainerCreated    = "services.dhs.notification.container.create"

	ExchangeContainerCreated      = "notfication.container.create"
	ExchangeContainerUpdated      = "notfication.container.update"
	ExchangeContainerDeleted      = "notfication.container.delete"
	ExchangeMetricCreated         = "notfication.metric.create"
	ExchangeMetricUpdated         = "notfication.metric.update"
	ExchangeMetricDeleted         = "notfication.metric.delete"
	ExchangeDataPolicyDeleted     = "notfication.data-policy.delete"
	ExchangeMetricDataRequest     = "global.metric.data_request"
	ExchangeMetricDataResponse    = "global.metric.data_response"
	ExchangeMetricsDataRequest    = "global.metrics.data_request"
	ExchangeMetricsDataResponse   = "global.metrics.data_response"
	ExchangeRTSMetricDataRequest  = "services.rts.metric.data_request"
	ExchangeRTSMetricDataResponse = "services.rts.metric.data_response"
)
