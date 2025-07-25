// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package cloudflarereceiver // import "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/cloudflarereceiver"

import (
	"errors"
	"fmt"
	"net"

	"go.opentelemetry.io/collector/config/configtls"
	"go.uber.org/multierr"
)

// Config holds all the parameters to start an HTTP server that can be sent logs from CloudFlare
type Config struct {
	Logs LogsConfig `mapstructure:"logs"`

	// prevent unkeyed literal initialization
	_ struct{}
}

type LogsConfig struct {
	Secret          string                  `mapstructure:"secret"`
	Endpoint        string                  `mapstructure:"endpoint"`
	TLS             *configtls.ServerConfig `mapstructure:"tls"`
	Attributes      map[string]string       `mapstructure:"attributes"`
	TimestampField  string                  `mapstructure:"timestamp_field"`
	TimestampFormat string                  `mapstructure:"timestamp_format"`
	Separator       string                  `mapstructure:"separator"`

	// prevent unkeyed literal initialization
	_ struct{}
}

var (
	errNoEndpoint = errors.New("an endpoint must be specified")
	errNoCert     = errors.New("tls was configured, but no cert file was specified")
	errNoKey      = errors.New("tls was configured, but no key file was specified")

	defaultTimestampField  = "EdgeStartTimestamp"
	defaultTimestampFormat = "rfc3339"
	defaultSeparator       = "."
)

func (c *Config) Validate() error {
	if c.Logs.Endpoint == "" {
		return errNoEndpoint
	}

	var errs error
	// Validate timestamp_format if provided
	if c.Logs.TimestampFormat != "" {
		switch c.Logs.TimestampFormat {
		case "unix", "unixnano", "rfc3339":
		default:
			errs = multierr.Append(errs, fmt.Errorf("invalid timestamp_format %q, must be one of: unix, unixnano, rfc3339", c.Logs.TimestampFormat))
		}
	}

	if c.Logs.TLS != nil {
		// Missing key
		if c.Logs.TLS.KeyFile == "" {
			errs = multierr.Append(errs, errNoKey)
		}

		// Missing cert
		if c.Logs.TLS.CertFile == "" {
			errs = multierr.Append(errs, errNoCert)
		}
	}

	_, _, err := net.SplitHostPort(c.Logs.Endpoint)
	if err != nil {
		errs = multierr.Append(errs, fmt.Errorf("failed to split endpoint into 'host:port' pair: %w", err))
	}

	return errs
}
