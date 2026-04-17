package db

import (
	"testing"

	"goto-db/internal/cli"
)

func TestResolveTarget_WithDBName(t *testing.T) {
	opts := &cli.Options{
		DBName:      "audit",
		Environment: "prod",
		Engine:      "postgres",
		User:        "testuser",
	}

	target, err := ResolveTarget(opts)
	if err != nil {
		t.Fatalf("ResolveTarget() error: %v", err)
	}

	expectedHost := "prod.primary.audit.db.viatorsystems.com"
	if target.Host != expectedHost {
		t.Errorf("Host = %q, want %q", target.Host, expectedHost)
	}
	if target.Port != 5432 {
		t.Errorf("Port = %d, want 5432", target.Port)
	}
	if target.LocalPort != 15432 {
		t.Errorf("LocalPort = %d, want 15432", target.LocalPort)
	}
	if target.Engine != "postgres" {
		t.Errorf("Engine = %q, want postgres", target.Engine)
	}
}

func TestResolveTarget_WithDBURL(t *testing.T) {
	opts := &cli.Options{
		DBURL:  "custom-db.example.com",
		Engine: "mysql",
		User:   "testuser",
	}

	target, err := ResolveTarget(opts)
	if err != nil {
		t.Fatalf("ResolveTarget() error: %v", err)
	}

	if target.Host != "custom-db.example.com" {
		t.Errorf("Host = %q, want custom-db.example.com", target.Host)
	}
	if target.Port != 3306 {
		t.Errorf("Port = %d, want 3306", target.Port)
	}
	if target.LocalPort != 13306 {
		t.Errorf("LocalPort = %d, want 13306", target.LocalPort)
	}
}

func TestResolveTarget_CustomLocalPort(t *testing.T) {
	opts := &cli.Options{
		DBName:      "booking",
		Environment: "rc",
		Engine:      "postgres",
		LocalPort:   9999,
		User:        "testuser",
	}

	target, err := ResolveTarget(opts)
	if err != nil {
		t.Fatalf("ResolveTarget() error: %v", err)
	}

	if target.LocalPort != 9999 {
		t.Errorf("LocalPort = %d, want 9999", target.LocalPort)
	}
}

func TestResolveTarget_MySQLEngine(t *testing.T) {
	opts := &cli.Options{
		DBName:      "orders",
		Environment: "prod",
		Engine:      "mysql",
		User:        "testuser",
	}

	target, err := ResolveTarget(opts)
	if err != nil {
		t.Fatalf("ResolveTarget() error: %v", err)
	}

	if target.Port != 3306 {
		t.Errorf("Port = %d, want 3306", target.Port)
	}
	if target.LocalPort != 13306 {
		t.Errorf("LocalPort = %d, want 13306", target.LocalPort)
	}
}
