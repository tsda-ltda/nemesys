package service

import (
	stdlog "log"
	"strconv"

	"github.com/fernandotsda/nemesys/shared/env"
)

type Service interface {
	Run()
	Done() <-chan error
	DispatchDone(error)
	GetServiceType() Type
	GetServiceNumber() int
	GetServiceIdent() string
	Close() error
}

type Type uint8

type NumberHandler struct {
	garbage []int
	n       int
}

type Tools struct {
	doneChs       []chan error
	ServiceNumber int
	ServiceIdent  string
	ServiceType   Type
}

const (
	Unknown Type = iota
	ServiceManager
	APIManager
	RTS
	DHS
	Alarm
	SNMP
	WS
)

var DefaultServiceNumber = NumberHandler{
	garbage: []int{},
	n:       1,
}

func (s *NumberHandler) Release(n int) {
	if n > s.n {
		panic("number was not created by ServiceNumber")
	}
	s.garbage = append(s.garbage, n)
}

func (s *NumberHandler) Get() (n int) {
	if len(s.garbage) > 0 {
		n = s.garbage[0]
		newGarbage := make([]int, len(s.garbage)-1)
		for i, v := range s.garbage {
			if i == 0 {
				continue
			}
			newGarbage[i-1] = v
		}
		s.garbage = newGarbage
		return n
	}
	n = s.n
	s.n++
	return n
}

func GetServiceName(t Type) string {
	switch t {
	case APIManager:
		return "API Manager"
	case RTS:
		return "Real time Service"
	case DHS:
		return "Data History Service"
	case Alarm:
		return "Alarm Service"
	case SNMP:
		return "SNMP service"
	case WS:
		return "Web Socket Service"
	case ServiceManager:
		return "Service Manager"
	default:
		panic("Unsupported service type")
	}
}

func GetServiceIdent(t Type, n int) (ident string) {
	switch t {
	case APIManager:
		ident = "api-manager-"
	case RTS:
		ident = "rts-"
	case DHS:
		ident = "dhs-"
	case Alarm:
		ident = "alarm-"
	case SNMP:
		ident = "snmp-"
	case WS:
		ident = "ws-"
	case ServiceManager:
		ident = "service-manager-"
	default:
		panic("Unsupported service type")
	}
	return ident + strconv.FormatInt(int64(n), 10)
}

func NewTools(t Type, n int) Tools {
	return Tools{
		doneChs:       make([]chan error, 0),
		ServiceNumber: n,
		ServiceType:   t,
		ServiceIdent:  GetServiceIdent(t, n),
	}
}

func (t *Tools) GetServiceNumber() int {
	return t.ServiceNumber
}

func (t *Tools) GetServiceIdent() string {
	return t.ServiceIdent
}

func (t *Tools) GetServiceType() Type {
	return t.ServiceType
}

func (st *Tools) DispatchDone(err error) {
	for _, v := range st.doneChs {
		v <- err
	}
}

func (st *Tools) Done() <-chan error {
	c := make(chan error)
	st.doneChs = append(st.doneChs, c)
	return c
}

func Start(t Type, instantiator func(serviceNumber int) Service, setups ...func(Service)) {
	err := env.LoadEnvFile()
	if err != nil {
		stdlog.Printf("Fail to load enviroment file, err:%s", err)
	}
	env.Init()

	number, err := registerService(t)
	if err != nil {
		stdlog.Fatalf("Fail to register service, err: %s", err)
		return
	}

	service := instantiator(number)
	for _, setup := range setups {
		setup(service)
	}
	service.Run()
}
