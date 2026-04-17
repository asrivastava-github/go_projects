package config

import "testing"

func TestNewConfig_Defaults(t *testing.T) {
	cfg := NewConfig("dev")

	if cfg.Environment != "dev" {
		t.Errorf("Environment = %q, want dev", cfg.Environment)
	}
	if cfg.Region != "eu-west-1" {
		t.Errorf("Region = %q, want eu-west-1", cfg.Region)
	}
	if cfg.Profile != "dev" {
		t.Errorf("Profile = %q, want dev (defaults to env)", cfg.Profile)
	}
}

func TestNewConfig_ProdRegion(t *testing.T) {
	cfg := NewConfig("prod")

	if cfg.Region != "us-east-1" {
		t.Errorf("Region = %q, want us-east-1 for prod", cfg.Region)
	}
	if cfg.Profile != "prod" {
		t.Errorf("Profile = %q, want prod", cfg.Profile)
	}
}

func TestNewConfig_ExplicitProfile(t *testing.T) {
	cfg := NewConfig("dev", "custom-profile")

	if cfg.Profile != "custom-profile" {
		t.Errorf("Profile = %q, want custom-profile", cfg.Profile)
	}
	if cfg.Environment != "dev" {
		t.Errorf("Environment = %q, want dev", cfg.Environment)
	}
}

func TestNewConfig_EmptyProfile(t *testing.T) {
	cfg := NewConfig("rc", "")

	if cfg.Profile != "rc" {
		t.Errorf("Profile = %q, want rc (empty string should fallback to env)", cfg.Profile)
	}
}

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		env     string
		wantErr bool
	}{
		{"valid env", "prod", false},
		{"valid env dev", "dev", false},
		{"empty env", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{Environment: tt.env}
			err := cfg.Validate()
			if tt.wantErr && err == nil {
				t.Error("Validate() expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("Validate() unexpected error: %v", err)
			}
		})
	}
}
