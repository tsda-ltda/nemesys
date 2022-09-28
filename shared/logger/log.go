package logger

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Return a production zap logger configurated.
func New(filelogPath string) (logger *zap.Logger, err error) {
	// create zap conf
	config := zap.NewProductionEncoderConfig()
	config.EncodeTime = zapcore.ISO8601TimeEncoder

	// encode logs to file as json
	fileEncoder := zapcore.NewJSONEncoder(config)

	// encode for console
	consoleEncoder := zapcore.NewConsoleEncoder(config)

	// open log file
	logFile, err := os.OpenFile(filelogPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}

	// sync for logfile
	writer := zapcore.AddSync(logFile)

	// default log level
	defaultLogLevel := zapcore.DebugLevel

	// create zap core
	core := zapcore.NewTee(
		zapcore.NewCore(fileEncoder, writer, defaultLogLevel),                        // log to file as json
		zapcore.NewCore(consoleEncoder, zapcore.AddSync(os.Stdout), defaultLogLevel), // log to console
	)
	return zap.New(core,
		zap.AddCaller(),
		zap.AddStacktrace(zapcore.ErrorLevel),
	), nil
}
