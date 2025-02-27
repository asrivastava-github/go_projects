# goto

`goto` binary is a simple script created from `main.go` that fetches EC2 instances to SSH into based on the `ROLE` and `Instance status` tags. It prompts you to choose the serial number of the instance to connect to and then executes the SSH command. The script uses `aws-vault` to retrieve the EC2 instance details.

```
 go build -o goto main.go
 sudo mv goto /usr/local/bin/\nchmod +x /usr/local/bin/goto
```

## Features

- Fetches EC2 instances based on `ROLE` and `Instance status` tags.
- Utilizes `aws-vault` for secure access to AWS credentials.
- Prompts user to select the instance to connect to.
- Executes SSH command to connect to the selected instance.

## Prerequisites

- `aws-vault` installed and configured with your AWS credentials.
- Go programming language installed.

## Usage

1. Ensure `aws-vault` is set up with your AWS credentials.
2. Run the `goto` binary.
3. Follow the prompts to select the EC2 instance you wish to connect to.
4. The script will execute the SSH command to connect to the selected instance.

## Example
goto <role-name> <environment>
