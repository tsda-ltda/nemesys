package amqp

const (
	QueueSNMPMetricDataReq      = "snmp_metric_data_req"
	QueueSNMPMetricsDataReq     = "snmp_metrics_data_req"
	QueueRTSMetricDataReq       = "rts_metric_data_req"
	QueueRTSMetricData          = "rts_metric_data"
	QueueDHSMetricsDataRes      = "dhs_metrics_data_res"
	QueueDHSMetricCreated       = "dhs_metric_created"
	QueueDHSContainerCreated    = "dhs_container_created"
	QueueAlarmCheckMetricAlarm  = "alarm_check_metric_alarm"
	QueueAlarmCheckMetricsAlarm = "alarm_check_metrics_alarm"
	QueueAlarmMetricAlarmed     = "alarm_metric_alarmed"
	QueueAlarmMetricsAlarmed    = "alarm_metrics_alarmed"

	ExchangeContainerCreated   = "container_created"    // fanout
	ExchangeContainerUpdated   = "container_updated"    // fanout
	ExchangeContainerDeleted   = "container_deleted"    // fanout
	ExchangeMetricCreated      = "metric_created"       // fanout
	ExchangeMetricUpdated      = "metric_updated"       // fanout
	ExchangeMetricDeleted      = "metric_deleted"       // fanout
	ExchangeDataPolicyDeleted  = "datapolicy_deleted"   // fanout
	ExchangeServiceLogs        = "logs"                 // fanout
	ExchangeServicesStatus     = "services_status"      // fanout
	ExchangeServiceRegisterReq = "register_service_req" // fanout
	ExchangeServiceRegisterRes = "register_service_res" // fanout
	ExchangeServiceUnregister  = "unregister_service"   // fanout
	ExchangeCheckMetricsAlarm  = "check_metrics_alarm"  // fanout
	ExchangeCheckMetricAlarm   = "check_metric_alarm"   // fanout
	ExchangeMetricsAlarmed     = "metrics_alarmed"      // fanout
	ExchangeMetricAlarmed      = "metric_alarmed"       // fanout

	ExchangeServicePing    = "ping"             // direct
	ExchangeServicePong    = "pong"             // direct
	ExchangeMetricDataReq  = "metric_data_req"  // direct
	ExchangeMetricsDataReq = "metrics_data_req" // direct
	ExchangeMetricDataRes  = "metric_data_res"  // direct
	ExchangeMetricsDataRes = "metrics_data_res" // direct
)
