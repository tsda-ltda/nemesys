package models

import (
	"errors"
	"time"

	"github.com/fernandotsda/nemesys/shared/amqp"
	"github.com/vmihailenco/msgpack/v5"
)

type AMQPMessage struct {
	// Type is the message type.
	Type amqp.MessageType `msgpack:"type"`
	// Metadata is a raw data sent as metadata.
	Metadata msgpack.RawMessage `msgpack:"metadata"`
	// Data is the message data.
	Data msgpack.RawMessage `msgpack:"data"`
}

type AMQPPlumber struct {
	Channels map[string]chan AMQPMessage
}

// GetWithTimeout will listen to the channel until it timeouts. Returns an error if timeouted.
func (p *AMQPPlumber) GetWithTimeout(channelKey string, timeout time.Duration) (AMQPMessage, error) {
	select {
	case res := <-p.Channels[channelKey]:
		return res, nil
	case <-time.After(timeout):
		return AMQPMessage{}, errors.New("pipe timeout")
	}
}
