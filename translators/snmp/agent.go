package snmp

import (
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/fernandotsda/nemesys/shared/models"
	"github.com/fernandotsda/nemesys/shared/types"
	g "github.com/gosnmp/gosnmp"
)

var ErrContainerNotExists = errors.New("container does not exists")

func (s *SNMPService) getContainerAgent(containerId int32, t types.ContainerType) (agent models.SNMPAgent, err error) {
	ctx := context.Background()

	r, err := s.cache.GetSNMPAgent(ctx, containerId)
	if err != nil {
		return agent, err
	}
	if r.Exists {
		return r.Agent, nil
	}

	switch t {
	case types.CTSNMPv2c:
		r, err := s.pg.GetSNMPv2cContainerProtocol(ctx, containerId)
		if err != nil {
			return agent, err
		}
		if !r.Exists {
			return agent, ErrContainerNotExists
		}

		agent = models.SNMPAgent{
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
		r, err := s.pg.GetFlexLegacyContainerProtocol(ctx, containerId)
		if err != nil {
			return agent, err
		}

		if !r.Exists {
			return agent, ErrContainerNotExists
		}

		agent = models.SNMPAgent{
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
		return agent, errors.New("unsupported container type: " + strconv.FormatInt(int64(t), 10))
	}
	err = s.cache.SetSNMPAgent(ctx, containerId, agent)
	return agent, err
}
