package models

import (
	"errors"
	"time"
)

type AMQPCorrelated[T any] struct {
	CorrelationId string
	RoutingKey    string
	Data          T
}

type AMQPPlumber struct {
	Channels map[string]chan []byte
}

// ListenWithTimeout will create and listen to the a plumber channel. Returns an error if timeouted.
func (p *AMQPPlumber) ListenWithTimeout(channelKey string, timeout time.Duration) ([]byte, error) {
	p.Channels[channelKey] = make(chan []byte)
	defer close(p.Channels[channelKey])
	defer delete(p.Channels, channelKey)
	select {
	case res := <-p.Channels[channelKey]:
		return res, nil
	case <-time.After(timeout):
		return nil, errors.New("pipe timeout")
	}
}
