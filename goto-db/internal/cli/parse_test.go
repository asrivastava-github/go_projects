package cli

import "testing"

func TestParse_ValidArgs(t *testing.T) {
	args := []string{"--db", "audit", "--env", "prod", "--user", "alice"}
	opts, err := Parse(args)
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}
	if opts.DBName != "audit" {
		t.Errorf("DBName = %q, want audit", opts.DBName)
	}
	if opts.Environment != "prod" {
		t.Errorf("Environment = %q, want prod", opts.Environment)
	}
	if opts.User != "alice" {
		t.Errorf("User = %q, want alice", opts.User)
	}
	if opts.Engine != "postgres" {
		t.Errorf("Engine = %q, want postgres (default)", opts.Engine)
	}
}

func TestParse_DBURLMode(t *testing.T) {
	args := []string{"--db-url", "custom.db.example.com", "--user", "bob"}
	opts, err := Parse(args)
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}
	if opts.DBURL != "custom.db.example.com" {
		t.Errorf("DBURL = %q, want custom.db.example.com", opts.DBURL)
	}
}

func TestParse_MissingDBAndURL(t *testing.T) {
	args := []string{"--env", "prod", "--user", "alice"}
	_, err := Parse(args)
	if err == nil {
		t.Error("Parse() expected error for missing --db and --db-url")
	}
}

func TestParse_MutuallyExclusive_DBAndURL(t *testing.T) {
	args := []string{"--db", "audit", "--db-url", "custom.db.com", "--user", "alice"}
	_, err := Parse(args)
	if err == nil {
		t.Error("Parse() expected error for --db and --db-url together")
	}
}

func TestParse_InvalidEngine(t *testing.T) {
	args := []string{"--db", "audit", "--engine", "sqlite", "--user", "alice"}
	_, err := Parse(args)
	if err == nil {
		t.Error("Parse() expected error for unsupported engine")
	}
}

func TestParse_MutuallyExclusive_AgentAndRefresh(t *testing.T) {
	args := []string{"--db", "audit", "--agent", "host1", "--refresh", "--user", "alice"}
	_, err := Parse(args)
	if err == nil {
		t.Error("Parse() expected error for --agent and --refresh together")
	}
}

func TestParse_RefreshOnly(t *testing.T) {
	args := []string{"--refresh", "--user", "alice"}
	opts, err := Parse(args)
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}
	if !opts.Refresh {
		t.Error("Refresh should be true")
	}
}

func TestParse_MySQLEngine(t *testing.T) {
	args := []string{"--db", "orders", "--engine", "mysql", "--user", "alice"}
	opts, err := Parse(args)
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}
	if opts.Engine != "mysql" {
		t.Errorf("Engine = %q, want mysql", opts.Engine)
	}
}
