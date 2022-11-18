package rdb

import (
	"fmt"
	"strconv"
)

// AuthSessionKey returns a RDB key for sessions.
func AuthSessionKey(session string) string {
	return "auth:sessions:" + session
}

// AuthReverseSessionKey returns a RDB key for user current session.
func AuthReverseSessionKey(userId int32) string {
	return "auth:users:sessions:" + strconv.FormatInt(int64(userId), 10)
}

// RDBCacheMetricIdContainerIdKey returns a RDB key for metric id and container id.
func RDBCacheMetricIdContainerIdKey(teamIdent string, contextIdent string, metricIdent string) string {
	return fmt.Sprintf("cache:metrics:%s_%s_%s", teamIdent, contextIdent, metricIdent)
}

// RDBCacheMetricDataKey returns a RDB key for metrics.
func RDBCacheMetricDataKey(metricId int64) string {
	return fmt.Sprintf("cache:metrics:%d:full", metricId)
}

func RDBCacheMetricEvExpressionKey(metricId int64) string {
	return fmt.Sprintf("cache:metrics:%d:evexpression", metricId)
}

func RDBCacheMetricDataPolicyId(metricId int64) string {
	return fmt.Sprintf("cache:metrics:%d:data-policy", metricId)
}
