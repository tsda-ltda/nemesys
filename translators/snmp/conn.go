package snmp

import (
	"context"
	"errors"
	"fmt"
	"time"

	g "github.com/gosnmp/gosnmp"
)

// Conn is the SNMP connection representation.
type Conn struct {
	// Id is the connection id.
	Id int
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
func (s *SNMPService) RegisterAgent(ctx context.Context, containerId int) error {
	// get agent configuration
	e, conf, err := s.pgConn.SNMPContainers.Get(ctx, containerId)
	if err != nil {
		return err
	}

	// check if exists
	if !e {
		return errors.New("snmp container not found")
	}

	c := &Conn{
		Agent: &g.GoSNMP{
			Target:             conf.Target,
			Port:               conf.Port,
			Community:          conf.Community,
			Transport:          conf.Transport,
			MaxOids:            conf.MaxOids,
			Timeout:            time.Millisecond * time.Duration(conf.Timeout),
			Retries:            conf.Retries,
			Version:            g.SnmpVersion(conf.Version),
			MsgFlags:           g.SnmpV3MsgFlags(conf.MsgFlags),
			ExponentialTimeout: false,
		},
		TTL:    time.Millisecond * time.Duration(conf.CacheDuration),
		Closed: make(chan any),
	}

	// connect to agent
	err = c.Agent.Connect()
	if err != nil {
		return err
	}

	// run ttl handler
	go c.RunTTL()
	c.OnClose = func(c *Conn) {
		// remove connection
		s.conns[c.Id] = nil
		s.Log.Debug("conn removed, addr: " + fmt.Sprint(c.Id))
	}

	// save connection
	s.conns[containerId] = c
	return nil
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
			print("a")
			c.Close()
			return
		}
	}
}
