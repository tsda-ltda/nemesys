package logger

import "go.uber.org/zap/zapcore"

func ErrField(err error) zapcore.Field {
	return zapcore.Field{
		Key:       "err",
		Type:      zapcore.ErrorType,
		Interface: err,
	}
}

func Int64Field(key string, value int64) zapcore.Field {
	return zapcore.Field{
		Key:     key,
		Type:    zapcore.Int64Type,
		Integer: value,
	}
}
func StringField(key string, value string) zapcore.Field {
	return zapcore.Field{
		Key:    key,
		Type:   zapcore.StringType,
		String: value,
	}
}
