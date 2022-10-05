package logger

import "go.uber.org/zap/zapcore"

func ErrField(err error) zapcore.Field {
	return zapcore.Field{
		Key:       "err",
		Type:      zapcore.ErrorType,
		Interface: err,
	}
}
