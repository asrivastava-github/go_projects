package awsauth

import (
	"log"
	"os"
	"os/exec"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sts"
)

// AreAWSCredentialsValid checks if AWS credentials are valid
func AreAWSCredentialsValid() bool {
	sess, err := session.NewSession(&aws.Config{})
	if err != nil {
		return false
	}

	svc := sts.New(sess)
	_, err = svc.GetCallerIdentity(&sts.GetCallerIdentityInput{})
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
