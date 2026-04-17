package cloudtrail

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	ct "github.com/aws/aws-sdk-go-v2/service/cloudtrail"
	"github.com/aws/aws-sdk-go-v2/service/cloudtrail/types"
)

func (c *Client) LookupByAction(ctx context.Context, action string, params LookupParams) ([]Event, error) {
	return c.lookup(ctx, types.LookupAttributeKeyEventName, action, params)
}

func (c *Client) LookupByUser(ctx context.Context, username string, params LookupParams) ([]Event, error) {
	return c.lookup(ctx, types.LookupAttributeKeyUsername, username, params)
}

func (c *Client) LookupByResource(ctx context.Context, resource string, params LookupParams) ([]Event, error) {
	return c.lookup(ctx, types.LookupAttributeKeyResourceName, resource, params)
}

type cloudTrailEventDetail struct {
	SourceIPAddress string `json:"sourceIPAddress"`
	ReadOnly        bool   `json:"readOnly"`
}

func (c *Client) lookup(ctx context.Context, key types.LookupAttributeKey, value string, params LookupParams) ([]Event, error) {
	now := time.Now().UTC()
	startTime := now.Add(-params.Duration)

	limit := params.Limit
	if limit <= 0 {
		limit = 50
	}

	input := &ct.LookupEventsInput{
		LookupAttributes: []types.LookupAttribute{
			{
				AttributeKey:   key,
				AttributeValue: aws.String(value),
			},
		},
		StartTime: aws.Time(startTime),
		EndTime:   aws.Time(now),
	}

	var events []Event
	var nextToken *string
	maxRetries := 5

	for {
		remaining := limit - len(events)
		if remaining <= 0 {
			break
		}
		pageSize := remaining
		if pageSize > 50 {
			pageSize = 50
		}
		input.MaxResults = aws.Int32(int32(pageSize))
		input.NextToken = nextToken

		var result *ct.LookupEventsOutput
		var err error

		for attempt := 0; attempt <= maxRetries; attempt++ {
			result, err = c.ct.LookupEvents(ctx, input)
			if err == nil {
				break
			}
			if attempt == maxRetries {
				return events, fmt.Errorf("CloudTrail lookup failed after retries: %w", err)
			}
			backoff := time.Duration(math.Pow(2, float64(attempt))) * 500 * time.Millisecond
			if backoff > 8*time.Second {
				backoff = 8 * time.Second
			}
			time.Sleep(backoff)
		}

		for _, e := range result.Events {
			event := Event{
				EventName:   deref(e.EventName),
				EventSource: deref(e.EventSource),
				Username:    deref(e.Username),
				EventID:     deref(e.EventId),
				AWSRegion:   c.region,
			}
			if e.EventTime != nil {
				event.Timestamp = *e.EventTime
			}

			if e.CloudTrailEvent != nil {
				var detail cloudTrailEventDetail
				if err := json.Unmarshal([]byte(*e.CloudTrailEvent), &detail); err == nil {
					event.SourceIP = detail.SourceIPAddress
					event.ReadOnly = detail.ReadOnly
				}
			}

			for _, r := range e.Resources {
				event.Resources = append(event.Resources, Resource{
					Type: deref(r.ResourceType),
					Name: deref(r.ResourceName),
				})
			}

			events = append(events, event)
		}

		if result.NextToken == nil || len(events) >= limit {
			break
		}
		nextToken = result.NextToken
	}

	return events, nil
}

func deref(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
