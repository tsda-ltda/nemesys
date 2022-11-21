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

	// get on cache
	r, err := s.cache.GetSNMPAgent(ctx, containerId)
	if err != nil {
		return agent, err
	}
	if r.Exists {
		return r.Agent, nil
	}

	// create container config
	switch t {
	case types.CTSNMPv2c:
		// get snmpv2c protocol configuration
		r, err := s.pgConn.SNMPv2cContainers.Get(ctx, containerId)
		if err != nil {
			return agent, err
		}

		// check if container exists
		if !r.Exists {
			return agent, ErrContainerNotExists
		}

		// fill agent
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
		// get flex legacy protocol configuration
		r, err := s.pgConn.FlexLegacyContainers.GetSNMPConfig(ctx, containerId)
		if err != nil {
			return agent, err
		}

		// check if container exists
		if !r.Exists {
			return agent, ErrContainerNotExists
		}

		// fill agent
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

	// save config on cache
	err = s.cache.SetSNMPAgent(ctx, containerId, agent)

	return agent, err
}
