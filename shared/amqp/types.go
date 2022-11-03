package amqp

import (
	"net/http"
	"strconv"
)

type MessageType uint8

const (
	Untyped MessageType = iota
	OK
	InternalError
	InvalidBody
	NotFound
	Failed
)

func ToMessageType(t string) MessageType {
	i, err := strconv.Atoi(t)
	if err != nil {
		return Untyped
	}
	return MessageType(i)
}

func FromMessageType(t MessageType) string {
	return strconv.Itoa(int(t))
}

func ParseToHttpStatus(t MessageType) int {
	switch t {
	case Untyped:
		return http.StatusInternalServerError
	case OK:
		return http.StatusOK
	case InvalidBody:
		return http.StatusInternalServerError
	case NotFound:
		return http.StatusNotFound
	case Failed:
		return http.StatusServiceUnavailable
	default:
		return http.StatusInternalServerError
	}
}
