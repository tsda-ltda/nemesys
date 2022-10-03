package logger

import "go.uber.org/zap/zapcore"

// Parse a string env value to zapcore Level.
func ParseLevelEnv(envValue string) zapcore.Level {
	switch envValue {
	case "debug":
		return zapcore.DebugLevel
	case "info":
		return zapcore.InfoLevel
	case "warn":
		return zapcore.WarnLevel
	case "error":
		return zapcore.ErrorLevel
	case "dpanic":
		return zapcore.PanicLevel
	case "panic":
		return zapcore.FatalLevel
	default:
		return zapcore.DebugLevel
	}
}
