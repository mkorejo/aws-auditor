package main

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/sts"
)

// STSAssumeRoleAPI defines the interface for the AssumeRole function.
// We use this interface to test the function using a mocked service.
type STSAssumeRoleAPI interface {
	AssumeRole(ctx context.Context,
		params *sts.AssumeRoleInput,
		optFns ...func(*sts.Options)) (*sts.AssumeRoleOutput, error)
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
