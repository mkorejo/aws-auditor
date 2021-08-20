package main

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/sts"
)

// STSAssumeRoleAPI defines the interface for the AssumeRole function.
// We use this interface to test the function using a mocked service.
type STSAssumeRoleAPI interface {
	AssumeRole(ctx context.Context,
		params *sts.AssumeRoleInput,
		optFns ...func(*sts.Options)) (*sts.AssumeRoleOutput, error)
}

func AssumeRole(client *sts.Client, accountId string, roleArn string) aws.Config {
	// Initialize STS session name
	sessionName := "sam-aws-auditor"

	var newConfig aws.Config

	// Setup input and perform AssumeRole
	stsInput := &sts.AssumeRoleInput{
		RoleArn:         &roleArn,
		RoleSessionName: &sessionName,
	}
	stsResult, err := TakeRole(context.TODO(), client, stsInput)
	if err != nil {
		log.Fatalln("Error performing AssumeRole:", err)
		return newConfig
	}

	// Initialize newConfig with temporary credentials from the AssumeRoleOutput
	newConfig, err = config.LoadDefaultConfig(context.TODO(), config.WithCredentialsProvider(credentials.StaticCredentialsProvider{
		Value: aws.Credentials{
			AccessKeyID:     *stsResult.Credentials.AccessKeyId,
			SecretAccessKey: *stsResult.Credentials.SecretAccessKey,
			SessionToken:    *stsResult.Credentials.SessionToken,
			Source:          "Auditor STS Credentials",
		},
	}))

	return newConfig
}

// TakeRole gets temporary security credentials to access resources.
// Inputs:
//     c is the context of the method call, which includes the AWS Region.
//     api is the interface that defines the method call.
//     input defines the input arguments to the service call.
// Output:
//     If successful, an AssumeRoleOutput object containing the result of the service call and nil.
//     Otherwise, nil and an error from the call to AssumeRole.
func TakeRole(c context.Context, api STSAssumeRoleAPI, input *sts.AssumeRoleInput) (*sts.AssumeRoleOutput, error) {
	return api.AssumeRole(c, input)
}
