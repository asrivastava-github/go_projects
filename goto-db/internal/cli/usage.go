package cli

import (
	"flag"
	"fmt"
)

func printUsage(fs *flag.FlagSet) {
	fmt.Println(`goto-db — SSH tunnel to databases via jump host + Jenkins agent

USAGE:
  goto-db --db <name> [--env <environment>] [--engine <type>] [OPTIONS]
  goto-db --db-url <full-hostname> [--engine <type>] [OPTIONS]

EXAMPLES:
  goto-db --db audit                          # prod.primary.audit.db.viatorsystems.com (postgres:5432)
  goto-db --db booking --env rc               # rc.primary.booking.db.viatorsystems.com
  goto-db --db audit --engine mysql           # prod.primary.audit.db.viatorsystems.com (mysql:3306)
  goto-db --db-url mydb.us-east-1.rds.amazonaws.com --local-port 15432
  goto-db --db audit --agent jenkins-agent70215c.prod.svc.ue1.viatorsystems.com
  goto-db --refresh                           # update cached Jenkins agent URL

OPTIONS:`)
	fs.PrintDefaults()
	fmt.Println(`
TUNNEL CHAIN:
  localhost:<local-port> → jump.ue1.viatorsystems.com → <jenkins-agent> → <db-host>:<db-port>`)
}
