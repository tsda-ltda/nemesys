package logger

import (
	"os"

	amqp "github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Logger struct {
	*zap.Logger
	conn       *amqp.Connection
	Config     Config
	amqpWriter *AMQPLoggerWriter
}

type Config struct {
	// Service name
	Service string
	// Enable level for logging messages in console.
	ConsoleLevel zapcore.Level
	// Enable level for broadcast messages in amqp server.
	BroadcastLevel zapcore.Level
}

// Return a production zap logger configurated to log to console and to amqp server.
func New(conn *amqp.Connection, config Config) (logger *Logger, err error) {
	// create amqp socket channel
	ch, err := conn.Channel()
	if err != nil {
		return nil, err
	}

	// exchange declare
	err = ch.ExchangeDeclare(
		"logs",   // name
		"fanout", // type
		true,     // durable
		false,    // auto-deleted
		false,    // internal
		false,    // no-wait
		nil,      // arguments
	)
	if err != nil {
		return nil, err
	}

	// create amqp logger writer
	amqpW := &AMQPLoggerWriter{
		ch: ch,
	}

	// create zap conf
	zapConfig := zap.NewProductionEncoderConfig()
	zapConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	// json encoder
	jsonEncoder := zapcore.NewJSONEncoder(zapConfig)
	jsonEncoder.AddString("serv", config.Service)

	// encode for console
	consoleEncoder := zapcore.NewConsoleEncoder(zapConfig)

	// create zap core
	core := zapcore.NewTee(
		zapcore.NewCore(jsonEncoder, zapcore.AddSync(amqpW), config.BroadcastLevel),      // log broadcast
		zapcore.NewCore(consoleEncoder, zapcore.AddSync(os.Stdout), config.ConsoleLevel), // log to console
	)

	// create zap logger
	z := zap.New(core,
		zap.AddCaller(),
		zap.AddStacktrace(zapcore.ErrorLevel),
	)

	// create logger
	logger = &Logger{
		conn:       conn,
		Logger:     z,
		amqpWriter: amqpW,
		Config:     config,
	}
	return logger, nil
}

func (l *Logger) Close() error {
	err := l.Logger.Sync()
	if err != nil {
		return err
	}
	return l.amqpWriter.ch.Close()
}
