package tlspool

import (
	"crypto/tls"
	"sync"
	"time"
)

type Config struct {
	// Network is the tls network.
	Network string
	// Timeout is the timeout to open a new
	// connection.
	Timeout time.Duration
	// MaxIdleCount is the maximum number of
	// idle connections.
	MaxIdleConn int
	// MaxIdleConnLifetime is the maximum time of
	// a connection as idle.
	MaxIdleConnLifetime time.Duration
	// MaxOpenConn is the maximum number of
	// open connections.
	MaxOpenConn int
	// Host is the host.
	Host string
	// Port is the host port.
	Port int
	// TLSConfig is the tls configuration.
	TLSConfig tls.Config
}

type TLSConnPool struct {
	// config is the tls pool configuration.
	config Config
	// idleConns is the current iddle connections.
	idleConns map[string]*TLSConn
	// connsOpen is the number of open connections.
	connsOpen int
	// requestCh is the connection request channel.
	requestCh chan *connReq
	// mu is the mutex used on put and get handlers.
	mu sync.Mutex
	// isClosed is the close status.
	isClosed bool
	// idleRemoverTicker is the ticker used to remove
	// idle connections.
	idleRemoverTicker *time.Ticker
}

func New(config Config) *TLSConnPool {
	p := &TLSConnPool{
		config:    config,
		idleConns: make(map[string]*TLSConn, config.MaxIdleConn),
		requestCh: make(chan *connReq),
		connsOpen: 0,
	}

	go p.handleConnReq()
	go p.removeIdleHandler()

	return p
}

// Put attempts to return a connection back to the pool.
// Closes the connection if the max idle connections is
// 0 or if the current number of idle connections is
// the maximum.
func (p *TLSConnPool) Put(c *TLSConn) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.isClosed {
		c.conn.Close()
		return
	}

	if p.config.MaxIdleConn == 0 || p.config.MaxIdleConn == len(p.idleConns) {
		c.conn.Close()
		c.pool.connsOpen--
	}

	p.idleConns[c.id] = c
}

// get retrieves tls connection. If no idle connection
// is available, attempts to create an new one, returning
// an timeout error if fail.
func (p *TLSConnPool) Get() (c *TLSConn, err error) {
	p.mu.Lock()
	if p.isClosed {
		p.mu.Unlock()
		return nil, ErrTLSConnPoolClosed

	}

	if len(p.idleConns) > 0 {
		for k, c := range p.idleConns {
			delete(p.idleConns, k)
			p.mu.Unlock()
			return c, nil
		}
	}

	if p.config.MaxOpenConn > 0 && p.connsOpen == p.config.MaxOpenConn {
		req := &connReq{
			connCh: make(chan *TLSConn),
			errCh:  make(chan error),
		}
		p.requestCh <- req
		p.mu.Unlock()

		select {
		case c := <-req.connCh:
			return c, nil
		case err := <-req.errCh:
			return nil, err
		}
	}

	p.connsOpen++
	p.mu.Unlock()

	c, err = p.newConn()
	if err != nil {
		p.mu.Lock()
		p.connsOpen--
		p.mu.Unlock()
	}

	return c, nil
}

// removeIdleHandler removes idle connection after the given
// max idle lifetime on pool config.
func (p *TLSConnPool) removeIdleHandler() {
	for range p.idleRemoverTicker.C {
		p.mu.Lock()
		for k := range p.idleConns {
			delete(p.idleConns, k)
		}
		p.mu.Unlock()
	}
}

// Close closes all idle connections
// and set the pool to closed moded.
func (p *TLSConnPool) Close() {
	p.mu.Lock()
	defer p.mu.Unlock()

	for k, c := range p.idleConns {
		c.conn.Close()
		delete(p.idleConns, k)
	}
	p.isClosed = true
}

// IsClosed returns the pool close status.
func (p *TLSConnPool) IsClosed() bool {
	return p.isClosed
}
