package auth

import (
	"strconv"
)

// SessionKey returns a RDB key for sessions.
func SessionKey(session string) string {
	return "sessions:" + session
}

// ReverseSessionKey returns a RDB key for user current session
func ReverseSessionKey(userId int) string {
	return "users:sessions:" + strconv.Itoa(userId)
}
