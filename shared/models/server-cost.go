package models

type ServerElements struct {
	Users                 int `json:"users"`
	Teams                 int `json:"teams"`
	Contexts              int `json:"contexts"`
	ContextualMetrics     int `json:"contextual-metrics"`
	BasicContainers       int `json:"basic-containers"`
	SNMPv2cContainers     int `json:"snmpv2c-containers"`
	FlexLegacyContainers  int `json:"flex-legacy-containers"`
	BasicMetrics          int `json:"basic-metrics"`
	SNMPv2cMetrics        int `json:"snmpv2c-metrics"`
	FlexLegacyMetrics     int `json:"flex-legacy-metrics"`
	CustomQueries         int `json:"custom-queries"`
	DataPolicies          int `json:"data-policies"`
	AlarmExpressions      int `json:"alarm-expressions"`
	AlarmProfiles         int `json:"alarm-profiles"`
	AlarmProfilesEmails   int `json:"alarm-profiles-emails"`
	AlarmCategories       int `json:"alarm-categories"`
	TrapCategoryRelations int `json:"trap-category-relations"`
	Refkeys               int `json:"ref-keys"`
	APIKeys               int `json:"api-keys"`
	InfluxDataPoints      int `json:"influx-data-points"`
	Requests              int `json:"requests"`
	RealtimeDataRequests  int `json:"realtime-data-requests"`
	DataHistoryRequests   int `json:"data-history-requests"`
}

type ServerBasePlan struct {
	Cost                  float64 `json:"cost"`
	Users                 int     `json:"users"`
	Teams                 int     `json:"teams"`
	Contexts              int     `json:"contexts"`
	ContextualMetrics     int     `json:"contextual-metrics"`
	BasicContainers       int     `json:"basic-containers"`
	SNMPv2cContainers     int     `json:"snmpv2c-containers"`
	FlexLegacyContainers  int     `json:"flex-legacy-containers"`
	BasicMetrics          int     `json:"basic-metrics"`
	SNMPv2cMetrics        int     `json:"snmpv2c-metrics"`
	FlexLegacyMetrics     int     `json:"flex-legacy-metrics"`
	CustomQueries         int     `json:"custom-queries"`
	DataPolicies          int     `json:"data-policies"`
	AlarmExpressions      int     `json:"alarm-expressions"`
	AlarmProfiles         int     `json:"alarm-profiles"`
	AlarmProfilesEmails   int     `json:"alarm-profiles-emails"`
	AlarmCategories       int     `json:"alarm-categories"`
	TrapCategoryRelations int     `json:"trap-category-relations"`
	Refkeys               int     `json:"ref-keys"`
	APIKeys               int     `json:"api-keys"`
	InfluxDataPoints      int     `json:"influx-data-points"`
	Requests              int     `json:"requests"`
	RealtimeDataRequests  int     `json:"realtime-data-requests"`
	DataHistoryRequests   int     `json:"data-history-requests"`
}

type ServerPriceTable struct {
	CoinType             string  `json:"coin-type" validate:"required,max=5"`
	User                 float64 `json:"user" validate:"min=0"`
	Team                 float64 `json:"team" validate:"min=0"`
	Context              float64 `json:"context" validate:"min=0"`
	ContextualMetric     float64 `json:"contextual-metric" validate:"min=0"`
	BasicContainer       float64 `json:"basic-container" validate:"min=0"`
	SNMPv2cContainer     float64 `json:"snmpv2c-container" validate:"min=0"`
	FlexLegacyContainer  float64 `json:"flex-legacy-container" validate:"min=0"`
	BasicMetric          float64 `json:"basic-metric" validate:"min=0"`
	SNMPv2cMetric        float64 `json:"snmpv2c-metric" validate:"min=0"`
	FlexLegacyMetric     float64 `json:"flex-legacy-metric" validate:"min=0"`
	CustomQuery          float64 `json:"custom-query" validate:"min=0"`
	DataPolicy           float64 `json:"data-policy" validate:"min=0"`
	AlarmExpression      float64 `json:"alarm-expression" validate:"min=0"`
	AlarmProfile         float64 `json:"alarm-profile" validate:"min=0"`
	AlarmProfileEmail    float64 `json:"alarm-profile-email" validate:"min=0"`
	AlarmCategory        float64 `json:"alarm-category" validate:"min=0"`
	TrapCategoryRelation float64 `json:"trap-category-relation" validate:"min=0"`
	Refkey               float64 `json:"ref-key" validate:"min=0"`
	APIKey               float64 `json:"api-key" validate:"min=0"`
	InfluxDataPoint      float64 `json:"influx-data-point" validate:"min=0"`
	Request              float64 `json:"request" validate:"min=0"`
	RealtimeDataRequest  float64 `json:"realtime-data-request" validate:"min=0"`
	DataHistoryRequests  float64 `json:"data-history-requests" validate:"min=0"`
}

type ServerCostResult struct {
	// GeneratedAt is the time in seconds when the result was generated.
	GeneratedAt    int64            `json:"generated-at"`
	PriceTable     ServerPriceTable `json:"price-table"`
	Elements       ServerElements   `json:"elements"`
	BasePlanCost   float64          `json:"base-plan-cost"`
	AdditionalCost float64          `json:"additional-cost"`
	TotalCost      float64          `json:"total-cost"`
}
