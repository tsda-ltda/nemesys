package db

import (
	"fmt"
	"strconv"
)

// RDBAuthSessionKey returns a RDB key for sessions.
func RDBAuthSessionKey(session string) string {
	return "auth:sessions:" + session
}

// RDBAuthReverseSessionKey returns a RDB key for user current session.
func RDBAuthReverseSessionKey(userId int32) string {
	return "auth:users:sessions:" + strconv.FormatInt(int64(userId), 10)
}

// RDBCacheMetricDataKey returns a RDB key for metrics.
func RDBCacheMetricDataKey(metricId int64) string {
	return "cache:metrics:" + strconv.FormatInt(metricId, 10)
}

// RDBCacheMetricIdContainerIdKey returns a RDB key for metric id and container id.
func RDBCacheMetricIdContainerIdKey(teamIdent string, contextIdent string, metricIdent string) string {
	return fmt.Sprintf("cache:metric_id:%s_%s_%s", teamIdent, contextIdent, metricIdent)
}

func RDBCacheMetricEvExpressionKey(metricId int64) string {
	return fmt.Sprintf("cache:metrics_evexpression:%d", metricId)
}
