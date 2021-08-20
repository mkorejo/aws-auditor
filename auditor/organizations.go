package main

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/organizations"
)

// GetActiveAccounts returns a map of Accounts IDs to Account Names for all active AWS accounts in the organization.
func GetActiveAccounts(config aws.Config) map[string]string {
	config.Region = "us-east-1"
	orgc := organizations.NewFromConfig(config)

	// aa is the map of Accounts IDs to Account Names
	aa := make(map[string]string)

	listAccountsInput := &organizations.ListAccountsInput{}
	listAccountsResult, err := orgc.ListAccounts(context.TODO(), listAccountsInput)
	if err != nil {
		log.Fatalln("Error retrieving accounts:", err)
		return aa
	}

	// For each active account, add an entry to the map
	for _, account := range listAccountsResult.Accounts {
		if account.Status == "ACTIVE" {
			aa[*account.Id] = *account.Name
		}
	}

	return aa
}

// GetManagementAccountID
func GetManagementAccountID(config aws.Config) string {
	config.Region = "us-east-1"
	orgc := organizations.NewFromConfig(config)

	// maid is the management account ID
	maid := ""

	describeOrgInput := &organizations.DescribeOrganizationInput{}
	describeOrgResult, err := orgc.DescribeOrganization(context.TODO(), describeOrgInput)
	if err != nil {
		log.Fatalln("Error describing organization:", err)
		return maid
	}

	maid = *describeOrgResult.Organization.MasterAccountId

	return maid
}
