package main

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

func TestSelectInstanceSingle(t *testing.T) {
	instances := []ec2types.Instance{
		{InstanceId: aws.String("i-abc123")},
	}
	got := selectInstance(instances)
	if *got.InstanceId != "i-abc123" {
		t.Errorf("selectInstance() returned %s, want i-abc123", *got.InstanceId)
	}
}

func TestSshToInstance_ExtractsFQDN(t *testing.T) {
	tests := []struct {
		name     string
		tags     []ec2types.Tag
		wantFQDN string
	}{
		{
			name: "FQDN tag present",
			tags: []ec2types.Tag{
				{Key: aws.String("Name"), Value: aws.String("my-server")},
				{Key: aws.String("FQDN"), Value: aws.String("server.example.com")},
			},
			wantFQDN: "server.example.com",
		},
		{
			name: "FQDN is first tag",
			tags: []ec2types.Tag{
				{Key: aws.String("FQDN"), Value: aws.String("first.example.com")},
			},
			wantFQDN: "first.example.com",
		},
		{
			name:     "no FQDN tag",
			tags:     []ec2types.Tag{{Key: aws.String("Name"), Value: aws.String("test")}},
			wantFQDN: "",
		},
		{
			name:     "no tags",
			tags:     nil,
			wantFQDN: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			instance := ec2types.Instance{Tags: tt.tags}
			got := extractFQDN(instance)
			if got != tt.wantFQDN {
				t.Errorf("extractFQDN() = %q, want %q", got, tt.wantFQDN)
			}
		})
	}
}

func TestParseArgs(t *testing.T) {
	tests := []struct {
		name       string
		args       []string
		wantRole   string
		wantEnv    string
		wantRegion string
		wantErr    bool
	}{
		{
			name:       "role and env only",
			args:       []string{"cmd", "webserver", "prod"},
			wantRole:   "webserver",
			wantEnv:    "prod",
			wantRegion: "us-east-1",
		},
		{
			name:       "role env and region",
			args:       []string{"cmd", "api", "dev", "eu-west-1"},
			wantRole:   "api",
			wantEnv:    "dev",
			wantRegion: "eu-west-1",
		},
		{
			name:    "too few args",
			args:    []string{"cmd"},
			wantErr: true,
		},
		{
			name:    "only one arg",
			args:    []string{"cmd", "role"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			role, env, region, err := parseArgs(tt.args)
			if tt.wantErr {
				if err == nil {
					t.Errorf("parseArgs() expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("parseArgs() unexpected error: %v", err)
				return
			}
			if role != tt.wantRole {
				t.Errorf("role = %q, want %q", role, tt.wantRole)
			}
			if env != tt.wantEnv {
				t.Errorf("env = %q, want %q", env, tt.wantEnv)
			}
			if region != tt.wantRegion {
				t.Errorf("region = %q, want %q", region, tt.wantRegion)
			}
		})
	}
}
