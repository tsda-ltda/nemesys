package logger

import (
	"os"

	"github.com/fernandotsda/nemesys/shared/amqp"
	"github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Logger struct {
	*zap.Logger
	conn       *amqp091.Connection
	WrapConfig Config
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
func New(conn *amqp091.Connection, config Config) (logger *Logger, err error) {
	// create amqp socket channel
	ch, err := conn.Channel()
	if err != nil {
		return nil, err
	}
	go func() {
		closed, canceled := amqp.OnChannelCloseOrCancel(ch)
		for {
			select {
			case err := <-closed:
				panic(err)
			case err := <-canceled:
				panic(err)
			}
		}
	}()

	// exchange declare
	err = ch.ExchangeDeclare(
		amqp.ExchangeServiceLogs, // name
		"fanout",                 // type
		true,                     // durable
		false,                    // auto-deleted
		false,                    // internal
		false,                    // no-wait
		nil,                      // arguments
	)
	if err != nil {
		return nil, err
	}

	amqpW := &AMQPLoggerWriter{
		ch: ch,
	}

	zapConfig := zap.NewProductionEncoderConfig()
	zapConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	jsonConfig := zap.NewProductionEncoderConfig()
	jsonConfig.EncodeTime = zapcore.RFC3339NanoTimeEncoder

	jsonEncoder := zapcore.NewJSONEncoder(jsonConfig)
	jsonEncoder.AddString("serv", config.Service)

	consoleEncoder := zapcore.NewConsoleEncoder(zapConfig)

	logfile, err := os.OpenFile(config.Service+".errors.logs.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}

	core := zapcore.NewTee(
		zapcore.NewCore(jsonEncoder, zapcore.AddSync(amqpW), config.BroadcastLevel),      // log broadcast
		zapcore.NewCore(consoleEncoder, zapcore.AddSync(os.Stdout), config.ConsoleLevel), // log to console
		zapcore.NewCore(consoleEncoder, zapcore.AddSync(logfile), zap.ErrorLevel),        // log to file
	)

	z := zap.New(core,
		zap.AddCaller(),
		zap.AddStacktrace(zapcore.ErrorLevel),
	)
	logger = &Logger{
		conn:       conn,
		Logger:     z,
		amqpWriter: amqpW,
		WrapConfig: config,
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
