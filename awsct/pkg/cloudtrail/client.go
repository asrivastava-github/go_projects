package cloudtrail

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/config"
	ct "github.com/aws/aws-sdk-go-v2/service/cloudtrail"
)

type Client struct {
	ct     *ct.Client
	region string
}

func NewClient(ctx context.Context, region, profile string) (*Client, error) {
	opts := []func(*config.LoadOptions) error{
		config.WithRegion(region),
	}
	if profile != "" {
		opts = append(opts, config.WithSharedConfigProfile(profile))
	}

	cfg, err := config.LoadDefaultConfig(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	return &Client{
		ct:     ct.NewFromConfig(cfg),
		region: region,
	}, nil
}
