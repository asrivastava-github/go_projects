package auditlog

import (
	"context"
	"fmt"
	"time"

	audit "google.golang.org/genproto/googleapis/cloud/audit"
	"google.golang.org/api/iterator"
	loggingpb "cloud.google.com/go/logging/apiv2/loggingpb"
	"google.golang.org/protobuf/proto"
)

func (c *Client) LookupByAction(ctx context.Context, method string, params LookupParams) ([]Event, error) {
	filter := fmt.Sprintf(`protoPayload.methodName="%s"`, method)
	return c.lookup(ctx, filter, params)
}

func (c *Client) LookupByUser(ctx context.Context, principal string, params LookupParams) ([]Event, error) {
	filter := fmt.Sprintf(`protoPayload.authenticationInfo.principalEmail="%s"`, principal)
	return c.lookup(ctx, filter, params)
}

func (c *Client) LookupByResource(ctx context.Context, resource string, params LookupParams) ([]Event, error) {
	filter := fmt.Sprintf(`protoPayload.resourceName:"%s"`, resource)
	return c.lookup(ctx, filter, params)
}

func (c *Client) lookup(ctx context.Context, extraFilter string, params LookupParams) ([]Event, error) {
	now := time.Now().UTC()
	startTime := now.Add(-params.Duration)

	limit := params.Limit
	if limit <= 0 {
		limit = 50
	}

	filter := fmt.Sprintf(
		`logName="projects/%s/logs/cloudaudit.googleapis.com%%2Factivity" AND %s AND timestamp>="%s" AND timestamp<="%s"`,
		c.projectID,
		extraFilter,
		startTime.Format(time.RFC3339),
		now.Format(time.RFC3339),
	)

	req := &loggingpb.ListLogEntriesRequest{
		ResourceNames: []string{fmt.Sprintf("projects/%s", c.projectID)},
		Filter:        filter,
		OrderBy:       "timestamp desc",
		PageSize:      int32(limit),
	}

	it := c.lc.ListLogEntries(ctx, req)

	var events []Event
	for len(events) < limit {
		entry, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return events, fmt.Errorf("error reading log entries: %w", err)
		}

		event := Event{
			Severity:  entry.Severity.String(),
			InsertID:  entry.InsertId,
			ProjectID: c.projectID,
		}

		if entry.Timestamp != nil {
			event.Timestamp = entry.Timestamp.AsTime()
		}

		if entry.Resource != nil {
			event.ResourceType = entry.Resource.Type
			if name, ok := entry.Resource.Labels["project_id"]; ok {
				event.ProjectID = name
			}
		}

		if pp := entry.GetProtoPayload(); pp != nil {
			var al audit.AuditLog
			if err := proto.Unmarshal(pp.Value, &al); err == nil {
				event.Method = al.GetMethodName()
				event.ServiceName = al.GetServiceName()
				event.ResourceName = al.GetResourceName()
				if ai := al.GetAuthenticationInfo(); ai != nil {
					event.Principal = ai.GetPrincipalEmail()
				}
				if rm := al.GetRequestMetadata(); rm != nil {
					event.CallerIP = rm.GetCallerIp()
				}
				if st := al.GetStatus(); st != nil && st.GetCode() != 0 {
					event.Status = fmt.Sprintf("%d", st.GetCode())
				} else {
					event.Status = "OK"
				}
			}
		}

		events = append(events, event)
	}

	return events, nil
}
