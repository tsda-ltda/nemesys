package snmp

import (
	"context"
	"errors"
	"time"

	g "github.com/gosnmp/gosnmp"
)

// Conn is the SNMP connection representation.
type Conn struct {
	// Id is the connection id.
	Id int32
	// TTL is the connection time to live.
	TTL time.Duration
	// Agent is the GoSNMP configuratation and connection.
	Agent *g.GoSNMP
	// Ticker is the TTL ticker controller.
	Ticker *time.Ticker
	// Closed is the channel to closed the connection.
	Closed chan any
	// OnClose is a callback used when connection is closed.
	OnClose func(c *Conn)
}

// RegisterConn register a connection.
func (s *SNMPService) RegisterAgent(ctx context.Context, containerId int32) (*Conn, error) {
	// get agent configuration
	e, conf, err := s.pgConn.SNMPContainers.Get(ctx, containerId)
	if err != nil {
		return nil, err
	}

	// check if exists
	if !e {
		return nil, errors.New("snmp container does not exists")
	}

	c := &Conn{
		Agent: &g.GoSNMP{
			Target:             conf.Target,
			Port:               uint16(conf.Port),
			Community:          conf.Community,
			Transport:          conf.Transport,
			MaxOids:            int(conf.MaxOids),
			Timeout:            time.Millisecond * time.Duration(conf.Timeout),
			Retries:            int(conf.Retries),
			Version:            g.SnmpVersion(conf.Version),
			MsgFlags:           g.SnmpV3MsgFlags(conf.MsgFlags),
			ExponentialTimeout: false,
		},
		TTL:    time.Millisecond * time.Duration(conf.CacheDuration),
		Closed: make(chan any),
		Id:     containerId,
	}

	// connect to agent
	err = c.Agent.Connect()
	if err != nil {
		return nil, err
	}

	// run ttl handler
	go c.RunTTL()
	c.OnClose = func(c *Conn) {
		// remove connection
		delete(s.conns, c.Id)
		s.Log.Debug("conn removed, addr: " + c.Agent.Target)
	}

	// save connection
	s.conns[c.Id] = c
	return c, nil
}

// Close closes agent connection and Closed chan.
func (c *Conn) Close() {
	c.Agent.Conn.Close()
	close(c.Closed)
	c.OnClose(c)
}

// Reset TTL ticker. Will panic if called before RunTTL.
func (c *Conn) Reset() {
	c.Ticker.Reset(c.TTL)
}

// RunTTL will set the connection ticker and close in the end.
func (c *Conn) RunTTL() {
	c.Ticker = time.NewTicker(c.TTL)
	defer c.Ticker.Stop()
	for {
		select {
		case <-c.Closed:
			return
		case <-c.Ticker.C:
			c.Close()
			return
		}
	}
}
