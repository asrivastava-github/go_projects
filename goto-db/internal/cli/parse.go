package cli

import (
	"flag"
	"fmt"
	"os/user"
)

func Parse(args []string) (*Options, error) {
	currentUser, _ := user.Current()
	defaultUser := ""
	if currentUser != nil {
		defaultUser = currentUser.Username
	}

	fs := flag.NewFlagSet("goto-db", flag.ContinueOnError)
	fs.Usage = func() { printUsage(fs) }

	opts := &Options{}
	fs.StringVar(&opts.DBName, "db", "", "Database short name (e.g., audit, booking)")
	fs.StringVar(&opts.Environment, "env", "prod", "Environment (prod, rc, int)")
	fs.StringVar(&opts.Engine, "engine", "postgres", "Database engine (postgres, mysql)")
	fs.StringVar(&opts.JenkinsAgent, "agent", "", "Jenkins agent hostname (overrides cached value)")
	fs.BoolVar(&opts.Refresh, "refresh", false, "Refresh and update the cached Jenkins agent URL")
	fs.IntVar(&opts.LocalPort, "local-port", 0, "Local port for the tunnel (default: same as remote DB port)")
	fs.StringVar(&opts.User, "user", defaultUser, "SSH username")
	fs.StringVar(&opts.DBURL, "db-url", "", "Fully qualified DB hostname (overrides --db and --env)")

	if err := fs.Parse(args); err != nil {
		return nil, err
	}

	if err := validate(opts); err != nil {
		fs.Usage()
		return nil, err
	}

	return opts, nil
}

func validate(opts *Options) error {
	if opts.DBName == "" && opts.DBURL == "" && !opts.Refresh {
		return fmt.Errorf("either --db or --db-url is required")
	}

	if opts.DBName != "" && opts.DBURL != "" {
		return fmt.Errorf("--db and --db-url are mutually exclusive")
	}

	if opts.Engine != "postgres" && opts.Engine != "mysql" {
		return fmt.Errorf("unsupported engine %q: must be postgres or mysql", opts.Engine)
	}

	if opts.JenkinsAgent != "" && opts.Refresh {
		return fmt.Errorf("--agent and --refresh are mutually exclusive")
	}

	if opts.User == "" {
		return fmt.Errorf("--user is required")
	}

	return nil
}
