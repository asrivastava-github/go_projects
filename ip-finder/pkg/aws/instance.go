package aws

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
)

type InstanceDetails struct {
	InstanceID   string
	Name         string
	InstanceType string
	State        string
	PrivateIP    string
	PublicIP     string
	Tags         map[string]string
}

type InstanceFinder struct {
	client *ec2.Client
}

func NewInstanceFinder(client *ec2.Client) *InstanceFinder {
	return &InstanceFinder{client: client}
}

func (f *InstanceFinder) GetDetails(ctx context.Context, instanceID string) (*InstanceDetails, error) {
	if instanceID == "" {
		return nil, nil
	}

	input := &ec2.DescribeInstancesInput{
		InstanceIds: []string{instanceID},
	}

	output, err := f.client.DescribeInstances(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to describe instance: %w", err)
	}

	if len(output.Reservations) == 0 || len(output.Reservations[0].Instances) == 0 {
		return nil, nil
	}

	instance := output.Reservations[0].Instances[0]
	details := &InstanceDetails{
		InstanceID:   aws.ToString(instance.InstanceId),
		InstanceType: string(instance.InstanceType),
		State:        string(instance.State.Name),
		PrivateIP:    aws.ToString(instance.PrivateIpAddress),
		PublicIP:     aws.ToString(instance.PublicIpAddress),
		Tags:         make(map[string]string),
	}

	for _, tag := range instance.Tags {
		details.Tags[aws.ToString(tag.Key)] = aws.ToString(tag.Value)
	}

	if name, ok := details.Tags["Name"]; ok {
		details.Name = name
	}

	return details, nil
}
