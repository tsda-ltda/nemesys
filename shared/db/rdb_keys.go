package db

import (
	"strconv"
)

// RDBAuthSessionKey returns a RDB key for sessions.
func RDBAuthSessionKey(session string) string {
	return "sessions:" + session
}

// RDBAuthReverseSessionKey returns a RDB key for user current session.
func RDBAuthReverseSessionKey(userId int) string {
	return "users:sessions:" + strconv.Itoa(userId)
}

// RDBRTSMetricKey returns a RDB key for metrics.
func RDBRTSMetricKey(metricId int) string {
	return "metrics:" + strconv.Itoa(metricId)
}
