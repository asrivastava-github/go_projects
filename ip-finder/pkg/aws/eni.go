package aws

import (
	"context"
	"fmt"
	"net"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

type ENIResult struct {
	NetworkInterfaceID string
	InterfaceType      string
	Description        string
	InstanceID         string
	Status             string
	PrivateIPs         []string
	SubnetID           string
	VPCID              string
	AvailabilityZone   string
	Tags               map[string]string
	FoundViaPrefix     bool
	MatchedPrefix      string
}

type ENIFinder struct {
	client *ec2.Client
	region string
}

func NewENIFinder(client *ec2.Client, region string) *ENIFinder {
	return &ENIFinder{
		client: client,
		region: region,
	}
}

func (f *ENIFinder) FindByIP(ctx context.Context, ip string) ([]ENIResult, error) {
	input := &ec2.DescribeNetworkInterfacesInput{
		Filters: []types.Filter{
			{
				Name:   aws.String("addresses.private-ip-address"),
				Values: []string{ip},
			},
		},
	}

	output, err := f.client.DescribeNetworkInterfaces(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to describe network interfaces: %w", err)
	}

	if len(output.NetworkInterfaces) > 0 {
		results := make([]ENIResult, 0, len(output.NetworkInterfaces))
		for _, eni := range output.NetworkInterfaces {
			results = append(results, f.toENIResult(eni))
		}
		return results, nil
	}

	return f.findByPrefixDelegation(ctx, ip)
}

func (f *ENIFinder) findByPrefixDelegation(ctx context.Context, ip string) ([]ENIResult, error) {
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return nil, nil
	}

	input := &ec2.DescribeNetworkInterfacesInput{
		Filters: []types.Filter{
			{
				Name:   aws.String("status"),
				Values: []string{"in-use"},
			},
		},
	}

	output, err := f.client.DescribeNetworkInterfaces(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to describe network interfaces: %w", err)
	}

	var results []ENIResult
	for _, eni := range output.NetworkInterfaces {
		for _, prefix := range eni.Ipv4Prefixes {
			if prefix.Ipv4Prefix == nil {
				continue
			}
			_, cidr, err := net.ParseCIDR(aws.ToString(prefix.Ipv4Prefix))
			if err != nil {
				continue
			}
			if cidr.Contains(parsedIP) {
				result := f.toENIResult(eni)
				result.FoundViaPrefix = true
				result.MatchedPrefix = aws.ToString(prefix.Ipv4Prefix)
				results = append(results, result)
				break
			}
		}
	}

	return results, nil
}

func (f *ENIFinder) toENIResult(eni types.NetworkInterface) ENIResult {
	result := ENIResult{
		NetworkInterfaceID: aws.ToString(eni.NetworkInterfaceId),
		InterfaceType:      string(eni.InterfaceType),
		Description:        aws.ToString(eni.Description),
		Status:             string(eni.Status),
		SubnetID:           aws.ToString(eni.SubnetId),
		VPCID:              aws.ToString(eni.VpcId),
		AvailabilityZone:   aws.ToString(eni.AvailabilityZone),
		Tags:               make(map[string]string),
		PrivateIPs:         make([]string, 0, len(eni.PrivateIpAddresses)),
	}

	if eni.Attachment != nil {
		result.InstanceID = aws.ToString(eni.Attachment.InstanceId)
	}

	for _, addr := range eni.PrivateIpAddresses {
		result.PrivateIPs = append(result.PrivateIPs, aws.ToString(addr.PrivateIpAddress))
	}

	for _, tag := range eni.TagSet {
		result.Tags[aws.ToString(tag.Key)] = aws.ToString(tag.Value)
	}

	return result
}
