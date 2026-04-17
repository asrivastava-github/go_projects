package auditlog

import "time"

type Event struct {
	Timestamp    time.Time  `json:"timestamp"`
	Method       string     `json:"method"`
	ServiceName  string     `json:"service_name"`
	Principal    string     `json:"principal"`
	CallerIP     string     `json:"caller_ip"`
	ResourceName string     `json:"resource_name"`
	ResourceType string     `json:"resource_type"`
	Status       string     `json:"status"`
	ProjectID    string     `json:"project_id"`
	Severity     string     `json:"severity"`
	InsertID     string     `json:"insert_id"`
}

type LookupParams struct {
	Duration  time.Duration
	ProjectID string
	Limit     int
}
