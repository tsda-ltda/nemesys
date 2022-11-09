package amqp

const (
	QueueRTSGetMetricData   = "services.rts.get_metric_data"
	QueueSNMPGetMetricData  = "services.snmp.get_metric_data"
	QueueSNMPGetMetricsData = "services.snmp.get_metrics_data"
	QueueRTSMetricData      = "services.rts.metric_data"
	QueueRTSMetricsData     = "services.rts.metrics_data"

	ExchangeNotifyContainer  = "services.notify.container"
	ExchangeNotifyMetric     = "services.notify.metric"
	ExchangeGetMetricData    = "services.get_metric_data"
	ExchangeGetMetricsData   = "services.get_metrics_data"
	ExchangeMetricData       = "services.metric_data"
	ExchangeMetricsData      = "services.metrics_data"
	ExchangeRTSGetMetricData = "services.rts.get_metric_data"
	ExchangeRTSMetricData    = "services.rts.metric_data"
)
