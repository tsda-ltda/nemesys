package snmp

import (
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/fernandotsda/nemesys/shared/types"
	g "github.com/gosnmp/gosnmp"
)

var ErrContainerNotExists = errors.New("container does not exists")

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
func (s *SNMPService) RegisterAgent(ctx context.Context, containerId int32, t types.ContainerType) (*Conn, error) {
	var agent *g.GoSNMP
	var ttl int32

	switch t {
	case types.CTSNMPv2c:
		// get snmpv2c protocol configuration
		r, err := s.pgConn.SNMPv2cContainers.Get(ctx, containerId)
		if err != nil {
			return nil, err
		}

		// check if container exists
		if !r.Exists {
			return nil, ErrContainerNotExists
		}

		// set ttl
		ttl = r.Container.CacheDuration

		// fill agent
		agent = &g.GoSNMP{
			Target:    r.Container.Target,
			Port:      uint16(r.Container.Port),
			Community: r.Container.Community,
			Transport: r.Container.Transport,
			Timeout:   time.Millisecond * time.Duration(r.Container.Timeout),
			MaxOids:   int(r.Container.MaxOids),
			Retries:   int(r.Container.Retries),
			Version:   g.Version2c,
		}
	case types.CTFlexLegacy:
		// get flex legacy protocol configuration
		r, err := s.pgConn.FlexLegacyContainers.GetSNMPConfig(ctx, containerId)
		if err != nil {
			return nil, err
		}

		// check if container exists
		if !r.Exists {
			return nil, ErrContainerNotExists
		}

		// set ttl
		ttl = r.Container.CacheDuration

		// fill agent
		agent = &g.GoSNMP{
			Target:    r.Container.Target,
			Port:      uint16(r.Container.Port),
			Community: r.Container.Community,
			Transport: r.Container.Transport,
			Timeout:   time.Millisecond * time.Duration(r.Container.Timeout),
			MaxOids:   int(r.Container.MaxOids),
			Retries:   int(r.Container.Retries),
			Version:   g.Version2c,
		}
	default:
		return nil, errors.New("unsupported container type: " + strconv.FormatInt(int64(t), 10))
	}

	// create connection
	c := &Conn{
		Id:     containerId,
		TTL:    time.Millisecond * time.Duration(ttl),
		Closed: make(chan any, 1),
		Agent:  agent,
	}

	// connect to agent
	err := c.Agent.Connect()
	if err != nil {
		return nil, err
	}

	// set no deadline
	c.Agent.Conn.SetDeadline(time.Time{})

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
	c.Closed <- struct{}{}
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
