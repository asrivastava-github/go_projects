package cloudtrail

import "time"

type Event struct {
	Timestamp   time.Time  `json:"timestamp"`
	EventName   string     `json:"event_name"`
	EventSource string     `json:"event_source"`
	Username    string     `json:"username"`
	SourceIP    string     `json:"source_ip"`
	ReadOnly    bool       `json:"read_only"`
	Resources   []Resource `json:"resources"`
	EventID     string     `json:"event_id"`
	AWSRegion   string     `json:"aws_region"`
}

type Resource struct {
	Type string `json:"type"`
	Name string `json:"name"`
}

type LookupParams struct {
	Duration time.Duration
	Region   string
	Limit    int
}
