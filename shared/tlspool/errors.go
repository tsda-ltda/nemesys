package tlspool

import "errors"

var (
	ErrTLSConnPoolClosed = errors.New("tls connection pool is closed")
	ErrConnReqTimeout    = errors.New("connection request timeout")
)
