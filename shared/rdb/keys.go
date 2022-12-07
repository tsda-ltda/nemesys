package rdb

import (
	"fmt"
	"strconv"
)

func AuthSessionKey(session string) string {
	return "auth:sessions:" + session
}

func AuthReverseSessionKey(userId int32) string {
	return "auth:users:sessions:" + strconv.FormatInt(int64(userId), 10)
}

func AuthAPIKeyKey(apikey string) string {
	return "auth:apikey:" + apikey
}

func AuthReverseAPIKeyKey(apikeyId int32) string {
	return "auth:apikey:id:" + strconv.FormatInt(int64(apikeyId), 10)
}

func CacheUserLimited(ip string) string {
	return "cache:user-limited:" + ip
}

func CacheMetricRequestByIdent(teamIdent string, contextIdent string, metricIdent string) string {
	return fmt.Sprintf("cache:metrics:%s_%s_%s", teamIdent, contextIdent, metricIdent)
}

func CacheMetricRequest(id int64) string {
	return fmt.Sprintf("cache:metrics:%d:metric-request", id)
}

func CacheMetricDataKey(metricId int64) string {
	return fmt.Sprintf("cache:metrics:%d:full", metricId)
}

func CacheMetricEvExpressionKey(metricId int64) string {
	return fmt.Sprintf("cache:metrics:%d:evexpression", metricId)
}

func CacheMetricDataPolicyIdKey(metricId int64) string {
	return fmt.Sprintf("cache:metrics:%d:data-policy", metricId)
}

func CacheGoSNMPConfigKey(containerId int32) string {
	return fmt.Sprintf("cache:containers:%d:go-snmp", containerId)
}

func CacheSNMPMetricKey(metricId int64) string {
	return fmt.Sprintf("cache:metrics:%d:snmp", metricId)
}

func CacheCustomQueryKey(cqId int32) string {
	return "cache:custom-query" + strconv.FormatInt(int64(cqId), 10)
}
func CacheCustomQueryByIdentKey(cqIdent string) string {
	return "cache:custom-query:" + cqIdent
}

func CacheMetricAddDataFormKey(refkey string) string {
	return "cache:metric-add-data-form:" + refkey
}

func CacheRTSMetricConfig(metricId int64) string {
	return "cache:metric:" + strconv.FormatInt(metricId, 10) + ":rts-config"
}

func CacheServerCostResultKey() string {
	return "cache:server-cost"
}
