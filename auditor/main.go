package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/remeh/sizedwaitgroup"
)

func handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
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

	// Assume credentials in the Audit account
	roleArn = "arn:aws:iam::" + Accounts["Audit"] + ":role/" + AWSAuditorRole
	currentConfig = AssumeRole(stsc, Accounts["Audit"], roleArn)

	// Fetch all active AWS accounts in the organization
	activeAccounts := GetActiveAccounts(currentConfig)
	log.Println(len(activeAccounts), "accounts in the organization with ACTIVE status")

	// Limit concurrency to 10 goroutines when querying multiple regions
	swg := sizedwaitgroup.New(10)

	// Iterate through the map of active AWS accounts
	managementAccountID := GetManagementAccountID(currentConfig)
	for accountID := range activeAccounts {
		// Skip the management account as this typically does not have the shared role
		if accountID == managementAccountID {
			continue
		}
		accountName := activeAccounts[accountID]

		// Assume credentials in the current account
		roleArn = "arn:aws:iam::" + accountID + ":role/" + AWSAuditorRole
		currentConfig = AssumeRole(stsc, accountID, roleArn)
		// log.Println("********** " + accountName + " (" + accountID + ") **********")

		for _, region := range Regions {
			currentConfig.Region = region
			swg.Add()
			go AuditConfig(currentConfig, accountID, accountName, &swg)
		}

		AuditIAMRoles(currentConfig, accountID, accountName)
		AuditIAMUsers(currentConfig, accountID, accountName)

		swg.Wait()
	}

	// Boilerplate handler logic for use with API Gateway
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
	lambda.Start(handler)
}
