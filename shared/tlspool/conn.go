package tlspool

import (
	"crypto/tls"
	"strconv"

	"github.com/fernandotsda/nemesys/shared/uuid"
)

type TLSConn struct {
	// id is the tls connection unique identifier.
	id string
	// pool is the tls connection pool.
	pool *TLSConnPool
	// conn is the tls connection.
	conn *tls.Conn
}

// Conn returns the underlying tls connection. You may
// not keep references of this connection.
func (c *TLSConn) Conn() *tls.Conn {
	return c.conn
}

// newConn attempts to create a new tls connection, dialing with
// the TLSConnPool configuration.
func (p *TLSConnPool) newConn() (c *TLSConn, err error) {
	tlsConn, err := tls.Dial(p.config.Network, p.config.Host+":"+strconv.FormatInt(int64(p.config.Port), 10), &p.config.TLSConfig)
	if err != nil {
		return nil, err
	}

	id, err := uuid.New()
	if err != nil {
		return nil, err
	}
	return &TLSConn{
		id:   id,
		conn: tlsConn,
		pool: p,
	}, nil
}
