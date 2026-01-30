package aws

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sts"
)

type CallerIdentity struct {
	Account string
	UserID  string
	Arn     string
}

func ValidateCredentials(ctx context.Context, region, profile string) (*CallerIdentity, error) {
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

	stsClient := sts.NewFromConfig(cfg)
	result, err := stsClient.GetCallerIdentity(ctx, &sts.GetCallerIdentityInput{})
	if err != nil {
		return nil, err
	}

	return &CallerIdentity{
		Account: deref(result.Account),
		UserID:  deref(result.UserId),
		Arn:     deref(result.Arn),
	}, nil
}

func deref(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func FormatCredentialError(profile string, err error) string {
	return fmt.Sprintf(`AWS credentials invalid or expired for profile '%s'

Error: %v

To fix, ensure aws-vault is installed and configured:
  brew install aws-vault  # if not installed

Then retry - ip-finder will automatically use aws-vault.`, profile, err)
}
