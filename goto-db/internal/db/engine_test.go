package db

import "testing"

func TestDefaultPort(t *testing.T) {
	tests := []struct {
		engine string
		want   int
	}{
		{"postgres", 5432},
		{"mysql", 3306},
		{"unknown", 5432},
	}
	for _, tt := range tests {
		t.Run(tt.engine, func(t *testing.T) {
			if got := DefaultPort(tt.engine); got != tt.want {
				t.Errorf("DefaultPort(%q) = %d, want %d", tt.engine, got, tt.want)
			}
		})
	}
}

func TestDefaultLocalPort(t *testing.T) {
	tests := []struct {
		engine string
		want   int
	}{
		{"postgres", 15432},
		{"mysql", 13306},
		{"unknown", 15432},
	}
	for _, tt := range tests {
		t.Run(tt.engine, func(t *testing.T) {
			if got := DefaultLocalPort(tt.engine); got != tt.want {
				t.Errorf("DefaultLocalPort(%q) = %d, want %d", tt.engine, got, tt.want)
			}
		})
	}
}

func TestDefaultClient(t *testing.T) {
	tests := []struct {
		engine string
		want   string
	}{
		{"postgres", "psql"},
		{"mysql", "mysql"},
		{"unknown", "psql"},
	}
	for _, tt := range tests {
		t.Run(tt.engine, func(t *testing.T) {
			if got := DefaultClient(tt.engine); got != tt.want {
				t.Errorf("DefaultClient(%q) = %q, want %q", tt.engine, got, tt.want)
			}
		})
	}
}
