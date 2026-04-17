package awsauth

import (
	"context"
	"log"
	"os"
	"os/exec"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sts"
)

// AreAWSCredentialsValid checks if AWS credentials are valid
func AreAWSCredentialsValid() bool {
	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		return false
	}

	stsClient := sts.NewFromConfig(cfg)
	_, err = stsClient.GetCallerIdentity(context.Background(), &sts.GetCallerIdentityInput{})
	return err == nil
}

// RenewCredentials attempts to renew AWS credentials using aws-vault
// Returns true if aws-vault was executed (process will exit)
func RenewCredentials(awsProfile string) bool {
	if !AreAWSCredentialsValid() {
		log.Printf("🔐 AWS credentials are missing or expired. Running `aws-vault exec %s`\n", awsProfile)
		// Preserve all command-line arguments
		cmdArgs := append([]string{"exec", awsProfile, "--", os.Args[0]}, os.Args[1:]...)
		cmd := exec.Command("aws-vault", cmdArgs...)

		cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
		err := cmd.Run()
		if err != nil {
			log.Fatalf("Error executing aws-vault: %v", err)
		}
		return true // aws-vault was executed
	}
	return false // no need to renew
}
