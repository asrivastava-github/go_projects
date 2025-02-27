package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/aws/aws-sdk-go/aws/credentials"

)

// Function to check if AWS credentials are valid
func areAWSCredentialsValid() bool {
	sess, err := session.NewSession(&aws.Config{})
	if err != nil {
		return false
	}

	svc := sts.New(sess)
	_, err = svc.GetCallerIdentity(&sts.GetCallerIdentityInput{})
	return err == nil
}

// Fetch instances based on role tag
func getEC2Instances(role string, awsRegion string) ([]*ec2.Instance, error) {
	// log.Printf("Fetching EC2 instances with Role: %s", role)

	sess, err := session.NewSession(&aws.Config{
    Region: aws.String(awsRegion),
    Credentials: credentials.NewEnvCredentials(),
	})

	if err != nil {
		return nil, fmt.Errorf("failed to create AWS session: %v", err)
	}
	
	svc := ec2.New(sess)
	// log.Printf("EC2 client created: %v", svc)
	filters := []*ec2.Filter{
		{
			Name:   aws.String("tag:Role"),
			Values: []*string{aws.String(role)},
		},
		{
			Name:   aws.String("instance-state-name"),
			Values: []*string{aws.String("running")},
		},
	}

	result, err := svc.DescribeInstances(&ec2.DescribeInstancesInput{Filters: filters})
	if err != nil {
		return nil, fmt.Errorf("failed to describe instances: %v", err)
	}

	var instances []*ec2.Instance
	for _, reservation := range result.Reservations {
		instances = append(instances, reservation.Instances...)
	}

	if len(instances) == 0 {
		log.Printf("No instances found for the given role: %v.", role)
	}

	return instances, nil
}

// Select instance if multiple found
func selectInstance(instances []*ec2.Instance) *ec2.Instance {
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

// Connect to selected instance via SSH
func sshToInstance(instance *ec2.Instance) {
	var fqdn string
	for _, tag := range instance.Tags {
		if *tag.Key == "FQDN" {
			fqdn = *tag.Value
			break
		}
	}

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

func main() {
	if len(os.Args) < 3 {
		log.Println("Usage: <binary> <role> <env> [aws_region]")
		os.Exit(1)
	}

	role := os.Args[1]
	env := os.Args[2]
	awsProfile := env
	awsRegion	:= "us-east-1"
	// log.Printf("%v", len(os.Args))
	// log.Printf("%v", os.Args)

	if len(os.Args) > 3 {
		awsRegion = os.Args[3]
	}

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
