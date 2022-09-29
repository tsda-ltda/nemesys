package auth

import "fmt"

// SessionKey returns a RDB key for sessions.
func SessionKey(session string) string {
	return fmt.Sprintf("sessions:%s", session)
}

func ReverseSessionKey(userId int) string {
	return fmt.Sprintf("users:sessions:%d", userId)
}
