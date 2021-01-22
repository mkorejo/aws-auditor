package main

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/sts"
)

var (
	// Accounts is a map of all AWS accounts we want to audit (Name to Account ID)
	Accounts = map[string]string{
		"Audit":           "984217156667",
		"Log archive":     "543705552769",
		"Network":         "797979728091",
		"Shared_Services": "625925655987",
	}

	// DefaultHTTPGetAddress is the site to GET
	DefaultHTTPGetAddress = "https://checkip.amazonaws.com"

	// ErrNoIP occurs when no IP is found in response
	ErrNoIP = errors.New("No IP in HTTP response")

	// ErrNon200Response occurs when non-200 status code is received
	ErrNon200Response = errors.New("Non-200 response found")
)

func handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	resp, err := http.Get(DefaultHTTPGetAddress)
	if err != nil {
		return events.APIGatewayProxyResponse{}, err
	}

	if resp.StatusCode != 200 {
		return events.APIGatewayProxyResponse{}, ErrNon200Response
	}

	ip, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return events.APIGatewayProxyResponse{}, err
	}

	if len(ip) == 0 {
		return events.APIGatewayProxyResponse{}, ErrNoIP
	}

	return events.APIGatewayProxyResponse{
		Body:       fmt.Sprintf("Hello, %v", string(ip)),
		StatusCode: 200,
	}, nil
}

func main() {
	// Load the default AWS configuration (~/.aws/config)
	initConfig, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatal("Configuration error:", err)
	}

	// Create an Amazon STS service client
	stsc := sts.NewFromConfig(initConfig)

	// roleArn and sessionName are sent to STS to perform AssumeRole in each account
	var roleArn, sessionName string
	var ra, sn *string

	// currentConfig represents the aws.Config for the current AssumeRole
	var currentConfig aws.Config

	// Initialize STS session name
	sessionName = "sam-aws-auditor"
	sn = &sessionName

	// Loop through all AWS accounts
	for accountName, accountID := range Accounts {
		roleArn = "arn:aws:iam::" + accountID + ":role/aws-controltower-ReadOnlyExecutionRole"
		log.Println("Auditing", accountName, "account with role", roleArn)

		ra = &roleArn

		// Setup STS input
		stsInput := &sts.AssumeRoleInput{
			RoleArn:         ra,
			RoleSessionName: sn,
		}

		// Perform AssumeRole
		stsResult, err := TakeRole(context.TODO(), stsc, stsInput)
		if err != nil {
			fmt.Println("Error performing AssumeRole:", err)
			return
		}

		// Initialize currentConfig with temporary credentials from the AssumeRoleOutput
		currentConfig, err = config.LoadDefaultConfig(context.TODO(), config.WithCredentialsProvider(credentials.StaticCredentialsProvider{
			Value: aws.Credentials{
				AccessKeyID:     *stsResult.Credentials.AccessKeyId,
				SecretAccessKey: *stsResult.Credentials.SecretAccessKey,
				SessionToken:    *stsResult.Credentials.SessionToken,
				Source:          "Auditor STS Credentials",
			},
		}))

		AuditIAM(currentConfig, accountName)
	}

	lambda.Start(handler)
}
