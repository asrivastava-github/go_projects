package db

import (
	"fmt"

	"goto-db/internal/cli"
)

const dbDomainSuffix = "db.viatorsystems.com"

type Target struct {
	Host      string
	Port      int
	LocalPort int
	Engine    string
}

func ResolveTarget(opts *cli.Options) (*Target, error) {
	remotePort := DefaultPort(opts.Engine)

	localPort := DefaultLocalPort(opts.Engine)
	if opts.LocalPort > 0 {
		localPort = opts.LocalPort
	}

	var host string
	if opts.DBURL != "" {
		host = opts.DBURL
	} else {
		host = fmt.Sprintf("%s.primary.%s.%s", opts.Environment, opts.DBName, dbDomainSuffix)
	}

	return &Target{
		Host:      host,
		Port:      remotePort,
		LocalPort: localPort,
		Engine:    opts.Engine,
	}, nil
}
