package influxdb

import (
	"context"
	"crypto/tls"
	"fmt"

	"github.com/fernandotsda/nemesys/shared/env"
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/domain"
)

type Client struct {
	influxdb2.Client

	// DefaultOrg is the default organization that will be used by de client.
	DefaultOrg *domain.Organization
}

// Connect connects to InfluxDB and create the default organixation if not exists, returning the client.
func Connect() (c Client, err error) {
	ctx := context.Background()

	// define protocol
	protocol := "http"

	// set TLS configuration
	var tlsConfig *tls.Config
	cer, err := tls.LoadX509KeyPair(env.InfluxTLSCertFilePath, env.InfluxTLSKeyFilePath)
	if err == nil {
		protocol = "https"
		tlsConfig = &tls.Config{
			Certificates: []tls.Certificate{cer},
		}
	}

	url := fmt.Sprintf("%s://%s:%s", protocol, env.InfluxHost, env.InfluxPort)
	client := influxdb2.NewClientWithOptions(url, env.InfluxToken,
		influxdb2.DefaultOptions().
			SetTLSConfig(tlsConfig),
	)

	// validate client connection health
	_, err = client.Health(ctx)
	if err != nil {
		return c, err
	}

	// get default organization
	orgApi := client.OrganizationsAPI()
	org, err := orgApi.FindOrganizationByName(ctx, env.InfluxOrg)
	if err != nil {
		// create organization
		org, err = orgApi.CreateOrganizationWithName(ctx, env.InfluxOrg)
		if err != nil {
			return c, err
		}
	}

	return Client{
		Client:     client,
		DefaultOrg: org,
	}, nil
}
