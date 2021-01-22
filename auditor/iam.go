package main

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
)

// AuditIAM performs read-only audit actions around IAM resources in an AWS account.
func AuditIAM(config aws.Config, accountName string) {
	iamc := iam.NewFromConfig(config)

	listUsersInput := &iam.ListUsersInput{}
	listRolesInput := &iam.ListRolesInput{}

	usersResult, err := iamc.ListUsers(context.TODO(), listUsersInput)
	if err != nil {
		fmt.Println("Got an error retrieving users:", err)
		return
	}

	rolesResult, err := iamc.ListRoles(context.TODO(), listRolesInput)
	if err != nil {
		fmt.Println("Got an error retrieving users:", err)
		return
	}

	if len(usersResult.Users) == 0 {
		log.Println("No users found in", accountName, "account")
	} else {
		for _, user := range usersResult.Users {
			fmt.Println(*user.UserName, " created on", *user.CreateDate)
		}
	}

	log.Println("# of roles:", len(rolesResult.Roles))
	for _, role := range rolesResult.Roles {
		// TODO - Only list roles that are not
		fmt.Println(*role.RoleName, " created on", *role.CreateDate)
	}
}
