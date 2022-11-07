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
	InvalidParse
	EvaluateFailed
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
	case InvalidParse:
		return http.StatusBadRequest
	case EvaluateFailed:
		return http.StatusBadRequest
	default:
		return http.StatusInternalServerError
	}
}

func GetMessage(t MessageType) string {
	switch t {
	case Failed:
		return "Fail to get metric's data."
	case InvalidParse:
		return "Fail to parse data to metric type. Check if the metric's type is correct."
	case EvaluateFailed:
		return "Fail to evaluate data with metric's evaluate expression."
	default:
		return ""
	}
}
