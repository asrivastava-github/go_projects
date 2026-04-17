package auditlog

import (
	"context"
	"fmt"

	logging "cloud.google.com/go/logging/apiv2"
)

type Client struct {
	lc        *logging.Client
	projectID string
}

func NewClient(ctx context.Context, projectID string) (*Client, error) {
	lc, err := logging.NewClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create logging client: %w", err)
	}

	return &Client{
		lc:        lc,
		projectID: projectID,
	}, nil
}

func (c *Client) Close() error {
	return c.lc.Close()
}
