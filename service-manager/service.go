package manager

import (
	"fmt"
	"sync"
	"time"

	"github.com/fernandotsda/nemesys/shared/service"
)

var muServiceNumber sync.Mutex

var (
	apiManagerN = service.DefaultServiceNumber
	dhsN        = service.DefaultServiceNumber
	alarmN      = service.DefaultServiceNumber
	snmpN       = service.DefaultServiceNumber
	wsN         = service.DefaultServiceNumber
	rtsN        = service.DefaultServiceNumber
)

type Service struct {
	// Name is the service name.
	Name string
	// Ident is the service ident.
	Ident string
	// Number is the service number.
	Number service.NumberType
	// Online is the online status;
	Online bool
	// LastPing is the time of the last ping
	LastPing time.Time
	// LostConnectionTime is the time of the connection lost.
	LostConnectionTime time.Time
	// Type is the service type.
	Type service.Type
}

func (s *ServiceManager) newService(t service.Type) Service {
	n := s.getServiceNumber(t)
	service := Service{
		Name:   service.GetServiceName(t),
		Ident:  service.GetServiceIdent(t, n),
		Number: n,
		Type:   t,
	}
	s.services = append(s.services, service)
	return service
}

func (s *ServiceManager) getServiceNumber(t service.Type) (n service.NumberType) {
	muServiceNumber.Lock()
	defer muServiceNumber.Unlock()
	switch t {
	case service.APIManager:
		n = apiManagerN.Get()
	case service.RTS:
		n = rtsN.Get()
	case service.Alarm:
		n = alarmN.Get()
	case service.DHS:
		n = dhsN.Get()
	case service.SNMP:
		n = snmpN.Get()
	case service.WS:
		n = wsN.Get()
	default:
		s.log.Fatal("Unsupported service type: " + fmt.Sprint(t))
		return 0
	}
	return n
}
