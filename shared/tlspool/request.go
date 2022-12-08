package tlspool

import (
	"time"
)

// connReq is channel wrapper to receive a tls connection
// or an error.
type connReq struct {
	// connCh is the channel to receive the tls connection.
	connCh chan *TLSConn
	// errCh is the channel to receive the error.
	errCh chan error
}

// handleConnReq listens to the request queue
// and attempts to fulfil any incoming requests.
func (p *TLSConnPool) handleConnReq() {
	for req := range p.requestCh {
		var (
			requestDone = false
			hasTimeout  = false
			timeoutChan = time.After(p.config.Timeout)
		)

		for {
			if requestDone || hasTimeout {
				break
			}
			select {
			case <-timeoutChan:
				hasTimeout = true
				req.errCh <- ErrConnReqTimeout
			default:
				p.mu.Lock()
				if len(p.idleConns) > 0 {
					for k, c := range p.idleConns {
						delete(p.idleConns, k)
						p.mu.Unlock()
						req.connCh <- c
						requestDone = true
						break
					}
				} else if p.connsOpen < p.config.MaxOpenConn {
					p.connsOpen++
					p.mu.Unlock()
					c, err := p.newConn()
					if err != nil {
						p.mu.Lock()
						p.connsOpen--
						p.mu.Unlock()
					} else {
						req.connCh <- c
						requestDone = true
					}
				} else {
					p.mu.Unlock()
				}
			}
		}
	}
}
