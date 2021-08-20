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
	"github.com/aws/aws-sdk-go-v2/service/sts"
)

var (
	// Accounts contains the IDs for the management and Audit accounts
	Accounts = map[string]string{
		"Management": "665735848255",
		"Audit":      "984217156667",
	}

	// Regions to audit
	Regions = [17]string{
		"us-east-1",
		"us-east-2",
		"us-west-1",
		"us-west-2",
		"ca-central-1",
		"eu-central-1",
		"eu-north-1",
		"eu-west-1",
		"eu-west-2",
		"eu-west-3",
		"ap-northeast-1",
		"ap-northeast-2",
		"ap-northeast-3",
		"ap-southeast-1",
		"ap-southeast-2",
		"ap-south-1",
		"sa-east-1",
	}

	// AWS Config
	// AWSConfigAggregatorName
	AWSConfigAggregatorName = "aws-controltower-GuardrailsComplianceAggregator"
	// AWSConfigRole must exist in every account with "ReadOnlyAccess" and "AWS_ConfigRole" policies attached.
	AWSConfigRole = "aws-controltower-ConfigRecorderRole"
	// AWSConfigSnapshotDeliveryFrequency
	// AWSConfigS3BucketName is the archival bucket in Log archive
	AWSConfigS3BucketName = "aws-controltower-logs-543705552769-us-east-1"
	// AWSConfigS3KeyPrefix
	// AWSConfigS3KMSKeyARN
	// AWSConfigSNSTopicAccount is the account which contains the SNS topic for delivery
	AWSConfigSNSTopicAccount = "984217156667"
	// AWSConfigSNSTopicName is the name of the SNS topic for delivery
	AWSConfigSNSTopicName = "aws-controltower-AllConfigNotifications"

	// Handler configuration
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

	// roleArn is sent to STS to perform AssumeRole in each account
	var roleArn string

	// currentConfig represents the aws.Config for the current AssumeRole
	var currentConfig aws.Config

	// Assume credentials in the Audit account to fetch all active AWS accounts in the organization
	roleArn = "arn:aws:iam::" + Accounts["Audit"] + ":role/aws-controltower-ReadOnlyExecutionRole"
	currentConfig = AssumeRole(stsc, Accounts["Audit"], roleArn)
	activeAccounts := GetActiveAccounts(currentConfig)
	log.Println(len(activeAccounts), "accounts in the organization with ACTIVE status")

	// Iterate through the map of active AWS accounts
	managementAccountID := GetManagementAccountID(currentConfig)
	for accountID := range activeAccounts {
		// Skip the management account as this typically does not have the shared role
		if accountID == managementAccountID {
			continue
		}
		accountName := activeAccounts[accountID]

		roleArn = "arn:aws:iam::" + accountID + ":role/aws-controltower-ReadOnlyExecutionRole"
		currentConfig = AssumeRole(stsc, accountID, roleArn)
		// log.Println("********** " + accountName + " (" + accountID + ") **********")

		for _, region := range Regions {
			currentConfig.Region = region
			AuditConfig(currentConfig, accountID, accountName)
		}
		AuditIAM(currentConfig, accountID, accountName)
	}

	lambda.Start(handler)
}
