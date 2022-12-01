package manager

import (
	"fmt"
	"sync"

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

func (s *ServiceManager) newService(t service.Type) service.ServiceStatus {
	n := s.getServiceNumber(t)
	service := service.ServiceStatus{
		Name:   service.GetServiceName(t),
		Ident:  service.GetServiceIdent(t, n),
		Number: n,
		Type:   t,
	}
	s.services = append(s.services, service)
	return service
}

func (s *ServiceManager) getServiceNumber(t service.Type) (n int) {
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
