package auditlog

import (
	"encoding/json"
	"testing"
	"time"
)

func TestEventJSONMarshaling(t *testing.T) {
	event := Event{
		Timestamp:    time.Date(2026, 4, 17, 10, 45, 3, 0, time.UTC),
		Method:       "compute.instances.delete",
		ServiceName:  "compute.googleapis.com",
		Principal:    "alice@example.com",
		CallerIP:     "10.0.1.50",
		ResourceName: "projects/my-project/zones/us-central1-a/instances/my-instance",
		ResourceType: "gce_instance",
		Status:       "OK",
		ProjectID:    "my-project",
		Severity:     "NOTICE",
		InsertID:     "abc123",
	}

	data, err := json.Marshal(event)
	if err != nil {
		t.Fatalf("Failed to marshal event: %v", err)
	}

	var decoded Event
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal event: %v", err)
	}

	if decoded.Method != event.Method {
		t.Errorf("Method = %q, want %q", decoded.Method, event.Method)
	}
	if decoded.Principal != event.Principal {
		t.Errorf("Principal = %q, want %q", decoded.Principal, event.Principal)
	}
	if decoded.ResourceName != event.ResourceName {
		t.Errorf("ResourceName = %q, want %q", decoded.ResourceName, event.ResourceName)
	}
}

func TestEventJSONFieldNames(t *testing.T) {
	event := Event{
		Method:      "compute.instances.delete",
		ServiceName: "compute.googleapis.com",
	}

	data, err := json.Marshal(event)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("Failed to unmarshal to map: %v", err)
	}

	expectedFields := []string{"method", "service_name", "principal", "caller_ip", "resource_name", "resource_type", "status", "project_id", "severity", "insert_id"}
	for _, field := range expectedFields {
		if _, ok := raw[field]; !ok {
			t.Errorf("Missing JSON field %q", field)
		}
	}
}
