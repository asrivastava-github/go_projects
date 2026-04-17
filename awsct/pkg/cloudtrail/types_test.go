package cloudtrail

import (
	"encoding/json"
	"testing"
	"time"
)

func TestEventJSONMarshaling(t *testing.T) {
	event := Event{
		Timestamp:   time.Date(2026, 4, 17, 10, 45, 3, 0, time.UTC),
		EventName:   "DeleteBucket",
		EventSource: "s3.amazonaws.com",
		Username:    "alice",
		SourceIP:    "10.0.1.50",
		ReadOnly:    false,
		Resources: []Resource{
			{Type: "AWS::S3::Bucket", Name: "my-bucket"},
		},
		EventID:   "abc123",
		AWSRegion: "us-east-1",
	}

	data, err := json.Marshal(event)
	if err != nil {
		t.Fatalf("Failed to marshal event: %v", err)
	}

	var decoded Event
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal event: %v", err)
	}

	if decoded.EventName != event.EventName {
		t.Errorf("EventName = %q, want %q", decoded.EventName, event.EventName)
	}
	if decoded.Username != event.Username {
		t.Errorf("Username = %q, want %q", decoded.Username, event.Username)
	}
	if len(decoded.Resources) != 1 {
		t.Fatalf("Resources length = %d, want 1", len(decoded.Resources))
	}
	if decoded.Resources[0].Name != "my-bucket" {
		t.Errorf("Resource name = %q, want %q", decoded.Resources[0].Name, "my-bucket")
	}
}

func TestEventJSONFieldNames(t *testing.T) {
	event := Event{
		EventName:   "RunInstances",
		EventSource: "ec2.amazonaws.com",
	}

	data, err := json.Marshal(event)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("Failed to unmarshal to map: %v", err)
	}

	expectedFields := []string{"event_name", "event_source", "username", "source_ip", "read_only", "resources", "event_id", "aws_region"}
	for _, field := range expectedFields {
		if _, ok := raw[field]; !ok {
			t.Errorf("Missing JSON field %q", field)
		}
	}
}
