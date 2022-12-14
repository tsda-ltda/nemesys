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

func (s *SNMP) getContainerAgent(containerId int32, t types.ContainerType) (agent models.SNMPv2cAgent, err error) {
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
		exists, container, err := s.pg.GetSNMPv2cContainerProtocol(ctx, containerId)
		if err != nil {
			return agent, err
		}
		if !exists {
			return agent, ErrContainerNotExists
		}

		agent = models.SNMPv2cAgent{
			Target:    container.Target,
			Port:      uint16(container.Port),
			Community: container.Community,
			Transport: container.Transport,
			Timeout:   time.Millisecond * time.Duration(container.Timeout),
			MaxOids:   int(container.MaxOids),
			Retries:   int(container.Retries),
			Version:   g.Version2c,
		}
	case types.CTFlexLegacy:
		exists, container, err := s.pg.GetFlexLegacyContainerProtocol(ctx, containerId)
		if err != nil {
			return agent, err
		}

		if !exists {
			return agent, ErrContainerNotExists
		}

		agent = models.SNMPv2cAgent{
			Target:    container.Target,
			Port:      uint16(container.Port),
			Community: container.Community,
			Transport: container.Transport,
			Timeout:   time.Millisecond * time.Duration(container.Timeout),
			MaxOids:   int(container.MaxOids),
			Retries:   int(container.Retries),
			Version:   g.Version2c,
		}
	default:
		return agent, errors.New("unsupported container type: " + strconv.FormatInt(int64(t), 10))
	}
	err = s.cache.SetSNMPAgent(ctx, containerId, agent)
	return agent, err
}
