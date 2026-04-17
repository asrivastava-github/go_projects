package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/aws-sdk-go-v2/service/sts"
)

// Function to check if AWS credentials are valid
func areAWSCredentialsValid() bool {
	ctx := context.Background()
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return false
	}

	stsClient := sts.NewFromConfig(cfg)
	_, err = stsClient.GetCallerIdentity(ctx, &sts.GetCallerIdentityInput{})
	return err == nil
}

// Fetch instances based on role tag
func getEC2Instances(role string, awsRegion string) ([]ec2types.Instance, error) {
	// log.Printf("Fetching EC2 instances with Role: %s", role)

	ctx := context.Background()
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(awsRegion))
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %v", err)
	}

	svc := ec2.NewFromConfig(cfg)
	// log.Printf("EC2 client created: %v", svc)
	filters := []ec2types.Filter{
		{
			Name:   aws.String("tag:Role"),
			Values: []string{role},
		},
		{
			Name:   aws.String("instance-state-name"),
			Values: []string{"running"},
		},
	}

	result, err := svc.DescribeInstances(ctx, &ec2.DescribeInstancesInput{Filters: filters})
	if err != nil {
		return nil, fmt.Errorf("failed to describe instances: %v", err)
	}

	var instances []ec2types.Instance
	for _, reservation := range result.Reservations {
		instances = append(instances, reservation.Instances...)
	}

	if len(instances) == 0 {
		log.Printf("No instances found for the given role: %v.", role)
	}

	return instances, nil
}

// Select instance if multiple found
func selectInstance(instances []ec2types.Instance) ec2types.Instance {
	if len(instances) == 1 {
		return instances[0]
	}

	log.Println("Multiple instances found. Select one:")
	for i, instance := range instances {
		for _, tag := range instance.Tags {
			if *tag.Key == "FQDN" {
				log.Printf("[%2d] %s (Instance ID: %s)\n", i+1, *tag.Value, *instance.InstanceId)
			}
		}
	}

	reader := bufio.NewReader(os.Stdin)
	log.Print("Enter selection: ")
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)

	choice := 0
	fmt.Sscanf(input, "%d", &choice)
	if choice < 1 || choice > len(instances) {
		log.Fatalf("Invalid selection")
	}

	return instances[choice-1]
}

// extractFQDN returns the FQDN tag value from an instance, or empty string if not found
func extractFQDN(instance ec2types.Instance) string {
	for _, tag := range instance.Tags {
		if *tag.Key == "FQDN" {
			return *tag.Value
		}
	}
	return ""
}

// Connect to selected instance via SSH
func sshToInstance(instance ec2types.Instance) {
	fqdn := extractFQDN(instance)

	if fqdn == "" {
		log.Fatal("No FQDN tag found for the selected instance")
	}

	log.Printf("Connecting to %s...\n", fqdn)
	cmd := exec.Command("ssh", fqdn)
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
	if err := cmd.Run(); err != nil {
		log.Fatalf("Failed to connect via SSH: %v", err)
	}
}

// parseArgs parses command-line arguments and returns role, env, region
func parseArgs(args []string) (role, env, region string, err error) {
	if len(args) < 3 {
		return "", "", "", fmt.Errorf("usage: <binary> <role> <env> [aws_region]")
	}
	role = args[1]
	env = args[2]
	region = "us-east-1"
	if len(args) > 3 {
		region = args[3]
	}
	return role, env, region, nil
}

func main() {
	role, env, awsRegion, err := parseArgs(os.Args)
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
	awsProfile := env

	// Check if AWS credentials exist & are valid
	if !areAWSCredentialsValid() {
		log.Printf("🔐 AWS credentials are missing or expired. Running `aws-vault exec` %v\n", awsProfile)
    // Preserve all command-line arguments
    cmdArgs := append([]string{"exec", awsProfile, "--", os.Args[0]}, os.Args[1:]...)
    cmd := exec.Command("aws-vault", cmdArgs...)

    cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
    err := cmd.Run()
    if err != nil {
        log.Fatalf("Error: %v", err)
    }
    os.Exit(0) // Exit after re-exec
}


	// Normal execution: AWS credentials should be valid now
	log.Printf("✅ Proceeding with EC2 query for role: %v in region %v & env: %v\n", role, awsRegion, env)
	instances, err := getEC2Instances(role, awsRegion)
	if err != nil {
		log.Fatalf("Error fetching instances: %v", err)
	}

	if len(instances) == 0 {
		log.Fatalf("No instances found for role %s in region %s", role, awsRegion)
	}

	instance := selectInstance(instances)
	sshToInstance(instance)
}
