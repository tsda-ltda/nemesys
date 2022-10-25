package snmp

import (
	"math"

	"github.com/gosnmp/gosnmp"
)

// Get fetch the OIDs's values. Returns an error only if an error is returned fo the SNMP Get request.
func (s *SNMPService) Get(c *Conn, oids []string) (res []gosnmp.SnmpPDU, err error) {
	// get agent
	a := c.Agent

	// oids buffer
	var oidsBuff []string
	if len(oids) >= a.MaxOids {
		oidsBuff = make([]string, a.MaxOids)
	} else {
		oidsBuff = make([]string, len(oids))
	}

	res = []gosnmp.SnmpPDU{}

	var i int
	for k := 0; k < int(math.Ceil(float64(len(oids))/float64(a.MaxOids))); k++ {
		// recalculate buffer
		r := len(oids) - k*a.MaxOids
		if r <= cap(oidsBuff) {
			oidsBuff = make([]string, r)
		}

		// get oids
		for j := 0; j < len(oidsBuff); j++ {
			oidsBuff[j] = oids[i]
			i++
		}

		// make request
		_res, _err := a.Get(oidsBuff)
		err = _err
		if err != nil {
			return res, err
		}

		// save response
		res = append(res, _res.Variables...)
	}
	return res, err
}
